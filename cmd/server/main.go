package main

import (
	"log"

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

	"github.com/labstack/echo/v4"
)

func main() {
	cfg := config.Load()

	db, err := database.NewMySQL(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

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
	articleService := article.NewService(articleRepo)
	articleHandler := article.NewHandler(articleService)

	blogRepo := blog.NewRepository(db)
	blogService := blog.NewService(blogRepo, articleRepo)
	blogHandler := blog.NewHandler(blogService)

	feedService := feed.NewService(recsysClient, articleRepo)
	feedHandler := feed.NewHandler(feedService)

	eventRepo := event.NewRepository(db)
	eventService := event.NewService(eventRepo)
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

	e.GET("/actuator/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "UP"})
	})

	authMiddleware := middleware.JWTAuth(jwtService)
	optionalAuth := middleware.OptionalJWTAuth(jwtService)

	v1 := e.Group("/v1")
	auth.RegisterRoutes(v1, authHandler, oauthHandler, authMiddleware)
	member.RegisterRoutes(v1, memberHandler, authMiddleware)
	article.RegisterRoutes(v1, articleHandler, authMiddleware, optionalAuth)
	blog.RegisterRoutes(v1, blogHandler, authMiddleware, optionalAuth)
	feed.RegisterRoutes(v1, feedHandler, authMiddleware, optionalAuth)
	event.RegisterRoutes(v1, eventHandler, authMiddleware, optionalAuth)
	link.RegisterRoutes(v1, linkHandler)
	meta.RegisterRoutes(v1, metaHandler)

	adminGroup := e.Group("/admin")
	admin.RegisterRoutes(adminGroup, adminHandler, authMiddleware, middleware.RequireRole("admin"))

	log.Printf("Starting server on :%s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
