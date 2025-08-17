package main

import (
	"context"
	"database/sql"
	"strings"
)

type DB interface {
	Driver() string
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type DBAdapter struct {
	inner      *sql.DB
	driverName string
}

func NewDBAdapter(inner *sql.DB, driverName string) *DBAdapter {
	return &DBAdapter{inner: inner, driverName: driverName}
}

func (d *DBAdapter) Driver() string { return d.driverName }

func (d *DBAdapter) Exec(query string, args ...any) (sql.Result, error) {
	return d.inner.Exec(d.rebind(query), args...)
}

func (d *DBAdapter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.inner.ExecContext(ctx, d.rebind(query), args...)
}

func (d *DBAdapter) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.inner.QueryContext(ctx, d.rebind(query), args...)
}

func (d *DBAdapter) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.inner.QueryRowContext(ctx, d.rebind(query), args...)
}

func (d *DBAdapter) rebind(query string) string {
	if d.driverName != "postgres" {
		return query
	}
	// convert '?' to $1, $2 ... outside of single/double quotes
	var b strings.Builder
	b.Grow(len(query) + 10)
	inSingle := false
	inDouble := false
	index := 1
	for i := 0; i < len(query); i++ {
		ch := query[i]
		switch ch {
		case '\'':
			if !inDouble { inSingle = !inSingle }
			b.WriteByte(ch)
		case '"':
			if !inSingle { inDouble = !inDouble }
			b.WriteByte(ch)
		case '?':
			if !inSingle && !inDouble {
				b.WriteByte('$')
				b.WriteString(intToString(index))
				index++
			} else {
				b.WriteByte(ch)
			}
		default:
			b.WriteByte(ch)
		}
	}
	return b.String()
}

func intToString(n int) string {
	// simple int -> string conversion without fmt to reduce deps here
	if n == 0 { return "0" }
	digits := [20]byte{}
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + (n % 10))
		n /= 10
	}
	return string(digits[i:])
}