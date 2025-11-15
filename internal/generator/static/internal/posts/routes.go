
package posts

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// RegisterRoutes registers all post routes with the given service
func RegisterRoutes(service Service, r chi.Router) {
	r.Route("/posts", func(r chi.Router) {
		r.Post("/", createPost(service))
		r.Get("/", listPosts(service))
		r.Get("/{post_id}", getPost(service))
		r.Put("/{post_id}", updatePost(service))
		r.Delete("/{post_id}", deletePost(service))
	})
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdatePostRequest struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

// getUserIDFromHeader extracts and validates the user ID from the X-User-ID header
func getUserIDFromHeader(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		jsonError(w, "Missing X-User-ID header", http.StatusBadRequest)
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		slog.Error("Invalid user ID", "error", err, "user_id", userIDStr)
		jsonError(w, "Invalid user ID", http.StatusBadRequest)
		return uuid.Nil, false
	}

	return userID, true
}

// createPost handles POST /posts
func createPost(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromHeader(w, r)
		if !ok {
			return
		}

		var req CreatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", "error", err)
			jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		post, err := service.CreatePost(r.Context(), userID, req.Title, req.Content)
		if err != nil {
			slog.Error("Failed to create post", "error", err)
			jsonError(w, "Failed to create post", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, post, http.StatusCreated)
	}
}

// getPost handles GET /posts/{post_id}
func getPost(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postIDStr := chi.URLParam(r, "post_id")
		postID, err := uuid.Parse(postIDStr)
		if err != nil {
			slog.Error("Invalid post_id", "error", err, "post_id", postIDStr)
			jsonError(w, "Invalid post_id", http.StatusBadRequest)
			return
		}

		post, err := service.GetPost(r.Context(), postID)
		if err == ErrPostNotFound {
			jsonError(w, "Post not found", http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("Failed to get post", "error", err)
			jsonError(w, "Failed to get post", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, post, http.StatusOK)
	}
}

// listPosts handles GET /posts
func listPosts(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			userIDStr = r.Header.Get("X-User-ID")
		}
		if userIDStr == "" {
			jsonError(w, "Missing user_id parameter or X-User-ID header", http.StatusBadRequest)
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			slog.Error("Invalid user ID", "error", err, "user_id", userIDStr)
			jsonError(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		postList, err := service.ListUserPosts(r.Context(), userID)
		if err != nil {
			slog.Error("Failed to list posts", "error", err, "user_id", userID)
			jsonError(w, "Failed to list posts", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, postList, http.StatusOK)
	}
}

// updatePost handles PUT /posts/{post_id}
func updatePost(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromHeader(w, r)
		if !ok {
			return
		}

		postIDStr := chi.URLParam(r, "post_id")
		postID, err := uuid.Parse(postIDStr)
		if err != nil {
			slog.Error("Invalid post_id", "error", err, "post_id", postIDStr)
			jsonError(w, "Invalid post_id", http.StatusBadRequest)
			return
		}

		var req UpdatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", "error", err)
			jsonError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		post, err := service.UpdatePost(r.Context(), postID, req.Title, req.Content)
		if err == ErrPostNotFound {
			jsonError(w, "Post not found", http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("Failed to update post", "error", err, "user_id", userID, "post_id", postID)
			jsonError(w, "Failed to update post", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, post, http.StatusOK)
	}
}

// deletePost handles DELETE /posts/{post_id}
func deletePost(service Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := getUserIDFromHeader(w, r)
		if !ok {
			return
		}

		postIDStr := chi.URLParam(r, "post_id")
		postID, err := uuid.Parse(postIDStr)
		if err != nil {
			slog.Error("Invalid post_id", "error", err, "post_id", postIDStr)
			jsonError(w, "Invalid post_id", http.StatusBadRequest)
			return
		}

		err = service.DeletePost(r.Context(), postID)
		if err == ErrPostNotFound {
			jsonError(w, "Post not found", http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("Failed to delete post", "error", err, "user_id", userID, "post_id", postID)
			jsonError(w, "Failed to delete post", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// jsonResponse writes a JSON response
func jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode response", "error", err)
	}
}

// jsonError writes a JSON error response
func jsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

