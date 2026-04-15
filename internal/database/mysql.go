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
	tlsMode := configureTLS()

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

func configureTLS() string {
	caPath := os.Getenv("DB_TLS_CA")
	if caPath == "" {
		return "skip-verify"
	}

	pem, err := os.ReadFile(caPath)
	if err != nil {
		log.Printf("DB TLS CA file not found (%s), falling back to skip-verify", caPath)
		return "skip-verify"
	}

	rootCertPool := x509.NewCertPool()
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		log.Println("failed to parse DB TLS CA certificate, falling back to skip-verify")
		return "skip-verify"
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
		log.Printf("failed to register TLS config: %v, falling back to skip-verify", err)
		return "skip-verify"
	}

	log.Println("MySQL TLS: using CA certificate verification")
	return "mysql-tls"
}
