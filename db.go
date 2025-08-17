package main

import (
	"context"
	"errors"
	"fmt"
)

func migrate(db DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			is_admin BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			source TEXT,
			published_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS feeds (
			id SERIAL PRIMARY KEY,
			url TEXT NOT NULL UNIQUE,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

func ensureDefaultAdmin(db DB, cfg Config) error {
	var exists int
	if err := db.QueryRowContext(context.Background(), "SELECT COUNT(1) FROM users WHERE is_admin = TRUE").Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	if cfg.DefaultAdmin == "" || cfg.DefaultAdminPwd == "" {
		return errors.New("missing default admin credentials in config")
	}
	hasher := NewPasswordHasher()
	hash, err := hasher.HashPassword(cfg.DefaultAdminPwd)
	if err != nil {
		return err
	}
	var id int64
	row := db.QueryRowContext(context.Background(), "INSERT INTO users (email, password_hash, is_admin) VALUES ($1, $2, TRUE) RETURNING id", cfg.DefaultAdmin, hash)
	if err := row.Scan(&id); err != nil { return err }
	return nil
}