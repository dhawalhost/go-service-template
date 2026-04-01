package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	gokiterrors "github.com/dhawalhost/gokit/errors"
	"github.com/dhawalhost/gokit/middleware"
	"github.com/dhawalhost/gokit/pagination"
	"github.com/dhawalhost/gokit/response"
	"github.com/dhawalhost/gokit/validator"

	"github.com/dhawalhost/go-service-template/internal/service"
)

// createRequest holds the data for creating a new example.
type createRequest struct {
	Name        string `json:"name"        validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

// updateRequest holds the data for updating an existing example.
type updateRequest struct {
	Name        string `json:"name"        validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

// List handles GET /api/v1/examples and returns paginated examples.
// Query parameters:
//   - page: page number (default: 1)
//   - page_size: number of items per page (default: 20)
//   - search: search term to filter by name (substring match, case-insensitive)
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.RequestIDFromContext(r.Context())
	tid := h.tenantID(r)

	params := pagination.ParseOffsetParams(r)

	items, total, err := h.svc.List(r.Context(), tid, service.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Search:   r.URL.Query().Get("search"),
	})
	if err != nil {
		h.log.Error("failed to list examples", zap.String("request_id", reqID), zap.Error(err))
		gokiterrors.WriteError(w, r, err)
		return
	}

	response.Paginated(w, r, items, params.ToPagination(total))
}

// Get handles GET /api/v1/examples/{id} and returns a single example.
// URL parameters:
//   - id: the ID of the example to retrieve
//
// Returns 404 if the example is not found or belongs to a different tenant.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.RequestIDFromContext(r.Context())
	tid := h.tenantID(r)
	id := chi.URLParam(r, "id")

	item, err := h.svc.Get(r.Context(), tid, id)
	if err != nil {
		h.log.Warn("failed to get example", zap.String("request_id", reqID), zap.String("id", id), zap.Error(err))
		gokiterrors.WriteError(w, r, err)
		return
	}

	response.Ok(w, r, item)
}

// Create handles POST /api/v1/examples and creates a new example.
// Request body should contain name and optional description fields.
// Returns 201 Created with the created example.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.RequestIDFromContext(r.Context())
	tid := h.tenantID(r)

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		gokiterrors.WriteError(w, r, gokiterrors.BadRequest("INVALID_BODY", err.Error()))
		return
	}

	if err := validator.Default.Validate(req); err != nil {
		gokiterrors.WriteError(w, r, gokiterrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}

	item, err := h.svc.Create(r.Context(), tid, service.CreateRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.log.Error("failed to create example", zap.String("request_id", reqID), zap.Error(err))
		gokiterrors.WriteError(w, r, err)
		return
	}

	response.Created(w, r, item)
}

// Update handles PUT /api/v1/examples/{id} and updates an existing example.
// URL parameters:
//   - id: the ID of the example to update
//
// Request body should contain name and optional description fields.
// Returns 404 if the example is not found or belongs to a different tenant.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.RequestIDFromContext(r.Context())
	tid := h.tenantID(r)
	id := chi.URLParam(r, "id")

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		gokiterrors.WriteError(w, r, gokiterrors.BadRequest("INVALID_BODY", err.Error()))
		return
	}

	if err := validator.Default.Validate(req); err != nil {
		gokiterrors.WriteError(w, r, gokiterrors.BadRequest("VALIDATION_ERROR", err.Error()))
		return
	}

	item, err := h.svc.Update(r.Context(), tid, id, service.UpdateRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		h.log.Error("failed to update example", zap.String("request_id", reqID), zap.String("id", id), zap.Error(err))
		gokiterrors.WriteError(w, r, err)
		return
	}

	response.Ok(w, r, item)
}

// Delete handles DELETE /api/v1/examples/{id} and deletes an example.
// URL parameters:
//   - id: the ID of the example to delete
//
// Returns 204 No Content on success.
// Returns 404 if the example is not found or belongs to a different tenant.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.RequestIDFromContext(r.Context())
	tid := h.tenantID(r)
	id := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), tid, id); err != nil {
		h.log.Error("failed to delete example", zap.String("request_id", reqID), zap.String("id", id), zap.Error(err))
		gokiterrors.WriteError(w, r, err)
		return
	}

	response.NoContent(w)
}
