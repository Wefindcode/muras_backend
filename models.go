package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// User

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserService struct {
	db             *sql.DB
	passwordHasher *PasswordHasher
}

func NewUserService(db *sql.DB, hasher *PasswordHasher) *UserService {
	return &UserService{db: db, passwordHasher: hasher}
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, email, password_hash, is_admin, created_at FROM users WHERE email = ?", email)
	u := &User{}
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*User, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, email, password_hash, is_admin, created_at FROM users WHERE id = ?", id)
	u := &User{}
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Create(ctx context.Context, email, password string, isAdmin bool) (*User, error) {
	hash, err := s.passwordHasher.HashPassword(password)
	if err != nil { return nil, err }
	res, err := s.db.ExecContext(ctx, "INSERT INTO users (email, password_hash, is_admin) VALUES (?, ?, ?)", email, hash, isAdmin)
	if err != nil { return nil, err }
	id, _ := res.LastInsertId()
	return s.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context) ([]*User, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, email, password_hash, is_admin, created_at FROM users ORDER BY id DESC")
	if err != nil { return nil, err }
	defer rows.Close()
	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.IsAdmin, &u.CreatedAt); err != nil { return nil, err }
		users = append(users, u)
	}
	return users, nil
}

// Post

type Post struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Source      *string    `json:"source,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type PostService struct { db *sql.DB }

func NewPostService(db *sql.DB) *PostService { return &PostService{db: db} }

func (s *PostService) Create(ctx context.Context, title, content string, source *string, publishedAt *time.Time) (*Post, error) {
	res, err := s.db.ExecContext(ctx, "INSERT INTO posts (title, content, source, published_at) VALUES (?, ?, ?, ?)", title, content, source, publishedAt)
	if err != nil { return nil, err }
	id, _ := res.LastInsertId()
	return s.GetByID(ctx, id)
}

func (s *PostService) Update(ctx context.Context, id int64, title, content string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE posts SET title = ?, content = ? WHERE id = ?", title, content, id)
	return err
}

func (s *PostService) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM posts WHERE id = ?", id)
	return err
}

func (s *PostService) GetByID(ctx context.Context, id int64) (*Post, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, title, content, source, published_at, created_at FROM posts WHERE id = ?", id)
	p := &Post{}
	if err := row.Scan(&p.ID, &p.Title, &p.Content, &p.Source, &p.PublishedAt, &p.CreatedAt); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PostService) List(ctx context.Context, limit, offset int) ([]*Post, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, title, content, source, published_at, created_at FROM posts ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil { return nil, err }
	defer rows.Close()
	var posts []*Post
	for rows.Next() {
		p := &Post{}
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Source, &p.PublishedAt, &p.CreatedAt); err != nil { return nil, err }
		posts = append(posts, p)
	}
	return posts, nil
}

// Feed

type Feed struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type FeedService struct { db *sql.DB }

func NewFeedService(db *sql.DB) *FeedService { return &FeedService{db: db} }

func (s *FeedService) Create(ctx context.Context, url string) (*Feed, error) {
	res, err := s.db.ExecContext(ctx, "INSERT INTO feeds (url, enabled) VALUES (?, 1)", url)
	if err != nil { return nil, err }
	id, _ := res.LastInsertId()
	return s.GetByID(ctx, id)
}

func (s *FeedService) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM feeds WHERE id = ?", id)
	return err
}

func (s *FeedService) List(ctx context.Context) ([]*Feed, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, url, enabled, created_at FROM feeds WHERE enabled = 1 ORDER BY id DESC")
	if err != nil { return nil, err }
	defer rows.Close()
	var feeds []*Feed
	for rows.Next() {
		f := &Feed{}
		if err := rows.Scan(&f.ID, &f.URL, &f.Enabled, &f.CreatedAt); err != nil { return nil, err }
		feeds = append(feeds, f)
	}
	return feeds, nil
}

func (s *FeedService) GetByID(ctx context.Context, id int64) (*Feed, error) {
	row := s.db.QueryRowContext(ctx, "SELECT id, url, enabled, created_at FROM feeds WHERE id = ?", id)
	f := &Feed{}
	if err := row.Scan(&f.ID, &f.URL, &f.Enabled, &f.CreatedAt); err != nil {
		return nil, err
	}
	return f, nil
}

var ErrNotFound = errors.New("not found")