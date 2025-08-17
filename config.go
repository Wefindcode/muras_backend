package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Port            string
	DatabaseURL     string
	DBDriver        string
	JWTSecret       string
	AllowCORS       bool
	DefaultAdmin    string
	DefaultAdminPwd string
}

func envOrDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func loadConfig() Config {
	cfg := Config{
		Port:            envOrDefault("PORT", "8080"),
		DatabaseURL:     envOrDefault("DATABASE_URL", "file:app.db?_foreign_keys=on"),
		DBDriver:        envOrDefault("DB_DRIVER", "sqlite"),
		JWTSecret:       envOrDefault("JWT_SECRET", "dev-secret-change-me"),
		AllowCORS:       envOrDefault("ALLOW_CORS", "true") == "true",
		DefaultAdmin:    envOrDefault("DEFAULT_ADMIN_EMAIL", "admin@example.com"),
		DefaultAdminPwd: envOrDefault("DEFAULT_ADMIN_PASSWORD", "admin123"),
	}
	return cfg
}

func openDatabase(cfg Config) (*sql.DB, string, error) {
	var driver string
	switch cfg.DBDriver {
	case "postgres", "pg", "postgresql":
		driver = "postgres"
	case "sqlite", "sqlite3", "":
		driver = "sqlite3"
	default:
		return nil, "", fmt.Errorf("unsupported DB_DRIVER: %s", cfg.DBDriver)
	}
	db, err := sql.Open(driver, cfg.DatabaseURL)
	if err != nil {
		return nil, "", fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, "", fmt.Errorf("ping db: %w", err)
	}
	return db, driver, nil
}