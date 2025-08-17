package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Auth

type AuthHandler struct {
	users      *UserService
	jwtManager *JWTManager
}

func NewAuthHandler(users *UserService, jwt *JWTManager) *AuthHandler {
	return &AuthHandler{users: users, jwtManager: jwt}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	u, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil || !h.users.passwordHasher.VerifyPassword(u.PasswordHash, req.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	tok, err := h.jwtManager.GenerateToken(u.ID, u.IsAdmin)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue token"})
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{Token: tok})
}

// Users

type UserHandler struct { users *UserService }

func NewUserHandler(s *UserService) *UserHandler { return &UserHandler{users: s} }

func (h *UserHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	for _, u := range users { u.PasswordHash = "" }
	writeJSON(w, http.StatusOK, users)
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

func (h *UserHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	u, err := h.users.Create(r.Context(), req.Email, req.Password, req.IsAdmin)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()}); return }
	u.PasswordHash = ""
	writeJSON(w, http.StatusCreated, u)
}

// Posts

type PostHandler struct { posts *PostService }

func NewPostHandler(s *PostService) *PostHandler { return &PostHandler{posts: s} }

func (h *PostHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	posts, err := h.posts.List(r.Context(), 50, 0)
	if err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	writeJSON(w, http.StatusOK, posts)
}

func (h *PostHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	p, err := h.posts.GetByID(r.Context(), id)
	if err != nil { writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"}); return }
	writeJSON(w, http.StatusOK, p)
}

type createPostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *PostHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	p, err := h.posts.Create(r.Context(), req.Title, req.Content, nil, nil)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()}); return }
	writeJSON(w, http.StatusCreated, p)
}

type updatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (h *PostHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var req updatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if err := h.posts.Update(r.Context(), id, req.Title, req.Content); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"updated": true})
}

func (h *PostHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.posts.Delete(r.Context(), id); err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()}); return }
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// Feeds

type FeedHandler struct { feeds *FeedService }

func NewFeedHandler(s *FeedService) *FeedHandler { return &FeedHandler{feeds: s} }

func (h *FeedHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.feeds.List(r.Context())
	if err != nil { writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()}); return }
	writeJSON(w, http.StatusOK, feeds)
}

type createFeedRequest struct { URL string `json:"url"` }

func (h *FeedHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req createFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	f, err := h.feeds.Create(r.Context(), req.URL)
	if err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()}); return }
	writeJSON(w, http.StatusCreated, f)
}

func (h *FeedHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.feeds.Delete(r.Context(), id); err != nil { writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()}); return }
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}