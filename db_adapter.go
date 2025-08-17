package main

import (
	"context"
	"database/sql"
)

type DB interface {
	Driver() string
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type DBAdapter struct {
	inner *sql.DB
}

func NewDBAdapter(inner *sql.DB) *DBAdapter {
	return &DBAdapter{inner: inner}
}

func (d *DBAdapter) Driver() string { return "postgres" }

func (d *DBAdapter) Exec(query string, args ...any) (sql.Result, error) {
	return d.inner.Exec(query, args...)
}

func (d *DBAdapter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.inner.ExecContext(ctx, query, args...)
}

func (d *DBAdapter) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.inner.QueryContext(ctx, query, args...)
}

func (d *DBAdapter) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.inner.QueryRowContext(ctx, query, args...)
}