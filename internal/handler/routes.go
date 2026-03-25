package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Pattern returns the base URL path for this service.
// RENAME_ME: change to your resource path e.g. "/api/v1/users".
func (h *Handler) Pattern() string { return "/api/v1/examples" }

// Router returns the chi sub-router with all routes registered.
func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}
