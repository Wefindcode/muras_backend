package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func waitForDatabase(db *DBAdapter, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			if err := db.inner.Ping(); err == nil {
				return nil
			}
		}
	}
}

func main() {
	cfg := loadConfig()

	dbConn, err := openDatabase(cfg)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer dbConn.Close()

	db := NewDBAdapter(dbConn)
	if err := waitForDatabase(db, 30*time.Second); err != nil {
		log.Fatalf("database not ready: %v", err)
	}

	if err := migrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	if err := ensureDefaultAdmin(db, cfg); err != nil {
		log.Fatalf("failed to ensure default admin: %v", err)
	}

	jwtManager := NewJWTManager(cfg.JWTSecret, 24*time.Hour)
	passwordHasher := NewPasswordHasher()
	feedService := NewFeedService(db)
	postService := NewPostService(db)
	userService := NewUserService(db, passwordHasher)

	// Start background feed worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go StartFeedWorker(ctx, feedService, postService, 10*time.Minute)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(jsonMiddleware)
	if cfg.AllowCORS {
		r.Use(corsMiddleware())
	}

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	// OpenAPI + Swagger UI
	r.Get("/openapi.yaml", ServeOpenAPI)
	r.Get("/docs", ServeSwaggerUI)

	// Auth
	authHandler := NewAuthHandler(userService, jwtManager)
	r.Post("/admin/login", authHandler.HandleLogin)

	// Posts
	postHandler := NewPostHandler(postService)
	r.Route("/posts", func(r chi.Router) {
		r.Get("/", postHandler.HandleList)
		r.Get("/{id}", postHandler.HandleGet)
		r.Group(func(r chi.Router) {
			r.Use(JWTAuthMiddleware(jwtManager))
			r.Use(AdminOnlyMiddleware(userService))
			r.Post("/", postHandler.HandleCreate)
			r.Put("/{id}", postHandler.HandleUpdate)
			r.Delete("/{id}", postHandler.HandleDelete)
		})
	})

	// Users
	userHandler := NewUserHandler(userService)
	r.Route("/users", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(JWTAuthMiddleware(jwtManager))
			r.Use(AdminOnlyMiddleware(userService))
			r.Get("/", userHandler.HandleList)
			r.Post("/", userHandler.HandleCreate)
		})
	})

	// Feeds (for the parser)
	feedHandler := NewFeedHandler(feedService)
	r.Route("/feeds", func(r chi.Router) {
		r.Get("/", feedHandler.HandleList)
		r.Group(func(r chi.Router) {
			r.Use(JWTAuthMiddleware(jwtManager))
			r.Use(AdminOnlyMiddleware(userService))
			r.Post("/", feedHandler.HandleCreate)
			r.Delete("/{id}", feedHandler.HandleDelete)
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("shutting down...")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	_ = srv.Shutdown(shutdownCtx)
}