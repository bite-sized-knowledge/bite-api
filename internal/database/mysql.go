package database

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bite-sized/bite-api/internal/config"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func NewMySQL(cfg *config.Config) (*sqlx.DB, error) {
	tlsMode := "skip-verify"
	if caPath := os.Getenv("DB_TLS_CA"); caPath != "" {
		rootCertPool := x509.NewCertPool()
		pem, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("read DB TLS CA: %w", err)
		}
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			return nil, fmt.Errorf("failed to parse DB TLS CA certificate")
		}
		if err := mysql.RegisterTLSConfig("mysql-tls", &tls.Config{
			RootCAs:            rootCertPool,
			InsecureSkipVerify: true, // skip hostname check (self-signed CN doesn't match Docker service name)
			VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
				cert, err := x509.ParseCertificate(rawCerts[0])
				if err != nil {
					return fmt.Errorf("parse peer cert: %w", err)
				}
				_, err = cert.Verify(x509.VerifyOptions{Roots: rootCertPool})
				return err
			},
		}); err != nil {
			return nil, fmt.Errorf("register TLS config: %w", err)
		}
		tlsMode = "mysql-tls"
		log.Println("MySQL TLS: using CA certificate verification")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=UTC&tls=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, tlsMode)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
