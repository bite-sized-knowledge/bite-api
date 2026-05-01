package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bite-sized/bite-api/internal/admin"
	"github.com/bite-sized/bite-api/internal/article"
	"github.com/bite-sized/bite-api/internal/auth"
	"github.com/bite-sized/bite-api/internal/blog"
	"github.com/bite-sized/bite-api/internal/config"
	"github.com/bite-sized/bite-api/internal/database"
	"github.com/bite-sized/bite-api/internal/event"
	"github.com/bite-sized/bite-api/internal/feed"
	"github.com/bite-sized/bite-api/internal/link"
	"github.com/bite-sized/bite-api/internal/member"
	"github.com/bite-sized/bite-api/internal/meta"
	"github.com/bite-sized/bite-api/internal/middleware"
	"github.com/bite-sized/bite-api/internal/recsys"
	"github.com/bite-sized/bite-api/pkg/email"
	jwtpkg "github.com/bite-sized/bite-api/pkg/jwt"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

const (
	readinessProbeTimeout = 2 * time.Second
	shutdownTimeout       = 30 * time.Second
)

func main() {
	cfg := config.Load()

	db, err := database.NewMySQL(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	jwtService := jwtpkg.NewService(cfg.JWTSecretKey, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
	emailClient := email.NewClient(cfg.ResendAPIKey, cfg.EmailFrom)
	recsysClient := recsys.NewClient(cfg.RecsysBaseURL, cfg.RecsysAPIKey)

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, jwtService, emailClient, cfg.AppBaseURL)
	authHandler := auth.NewHandler(authService, cfg.JWTRefreshExpiry)

	memberRepo := member.NewRepository(db)
	memberService := member.NewService(memberRepo, jwtService)
	memberHandler := member.NewHandler(memberService, cfg.JWTRefreshExpiry)

	oauthRepo := auth.NewOAuthRepository(db)
	oauthService := auth.NewOAuthService(cfg, oauthRepo, memberRepo, jwtService)
	oauthHandler := auth.NewOAuthHandler(oauthService, cfg.JWTRefreshExpiry)

	articleRepo := article.NewRepository(db)
	articleService := article.NewService(articleRepo, recsysClient, cfg.RecsysSearchEnabled)
	articleHandler := article.NewHandler(articleService)

	blogRepo := blog.NewRepository(db)
	blogService := blog.NewService(blogRepo, articleRepo)
	blogHandler := blog.NewHandler(blogService)

	feedService := feed.NewService(recsysClient, articleRepo)
	feedHandler := feed.NewHandler(feedService)

	eventRepo := event.NewRepository(db)
	eventService := event.NewService(eventRepo, recsysClient)
	eventHandler := event.NewHandler(eventService)

	linkService := link.NewService(articleRepo)
	linkHandler := link.NewHandler(linkService)

	metaRepo := meta.NewRepository(db)
	metaHandler := meta.NewHandler(metaRepo)

	adminService := admin.NewService(articleRepo, blogRepo)
	adminHandler := admin.NewHandler(adminService)

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	e.Use(middleware.SecureHeaders())

	// Liveness — 프로세스 살아있음만 확인. 재시작 트리거 용도.
	e.GET("/actuator/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "UP"})
	})

	// Readiness — 트래픽 받을 준비. DB는 hard dependency,
	// recsys는 down이어도 MySQL FULLTEXT fallback이 있어 soft dependency로 보고만 함.
	e.GET("/readyz", readyzHandler(db, cfg.RecsysBaseURL))

	authMiddleware := middleware.JWTAuth(jwtService)
	optionalAuth := middleware.OptionalJWTAuth(jwtService)
	lazyGuest := middleware.LazyGuest(memberService.IssueGuestForDevice, cfg.JWTRefreshExpiry)

	v1 := e.Group("/v1")
	auth.RegisterRoutes(v1, authHandler, oauthHandler, authMiddleware)
	member.RegisterRoutes(v1, memberHandler, authMiddleware, optionalAuth, lazyGuest)
	article.RegisterRoutes(v1, articleHandler, authMiddleware, optionalAuth, lazyGuest)
	blog.RegisterRoutes(v1, blogHandler, authMiddleware, optionalAuth, lazyGuest)
	feed.RegisterRoutes(v1, feedHandler, authMiddleware, optionalAuth)
	event.RegisterRoutes(v1, eventHandler, authMiddleware, optionalAuth)
	link.RegisterRoutes(v1, linkHandler)
	meta.RegisterRoutes(v1, metaHandler)

	adminGroup := e.Group("/admin")
	admin.RegisterRoutes(adminGroup, adminHandler, authMiddleware, middleware.RequireRole("admin"))

	go func() {
		log.Printf("Starting server on :%s", cfg.Port)
		if err := e.Start(":" + cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("received signal %s, shutting down (timeout %s)", sig, shutdownTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	if err := db.Close(); err != nil {
		log.Printf("db close error: %v", err)
	}
	log.Println("server stopped cleanly")
}

func readyzHandler(db *sqlx.DB, recsysBaseURL string) echo.HandlerFunc {
	probeClient := &http.Client{Timeout: readinessProbeTimeout}
	return func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), readinessProbeTimeout)
		defer cancel()

		checks := map[string]string{}
		ready := true

		if err := db.PingContext(ctx); err != nil {
			checks["db"] = "down: " + err.Error()
			ready = false
		} else {
			checks["db"] = "ok"
		}

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, recsysBaseURL+"/health", nil)
		resp, err := probeClient.Do(req)
		switch {
		case err != nil:
			checks["recsys"] = "degraded: " + err.Error()
		case resp.StatusCode >= 500:
			resp.Body.Close()
			checks["recsys"] = "degraded: status " + resp.Status
		default:
			resp.Body.Close()
			checks["recsys"] = "ok"
		}

		status := http.StatusOK
		if !ready {
			status = http.StatusServiceUnavailable
		}
		return c.JSON(status, map[string]any{"ready": ready, "checks": checks})
	}
}
