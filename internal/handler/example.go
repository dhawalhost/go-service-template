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

type createRequest struct {
Name        string `json:"name"        validate:"required,min=1,max=255"`
Description string `json:"description" validate:"max=1000"`
}

type updateRequest struct {
Name        string `json:"name"        validate:"omitempty,min=1,max=255"`
Description string `json:"description" validate:"omitempty,max=1000"`
}

// tenantID extracts the tenant ID from context, defaulting to "default".
func tenantID(r *http.Request) string {
tid, ok := middleware.TenantIDFromContext(r.Context())
if !ok {
return "default"
}
return tid
}

// List handles GET /api/v1/examples
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
reqID := middleware.RequestIDFromContext(r.Context())
tid := tenantID(r)

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

// Get handles GET /api/v1/examples/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
reqID := middleware.RequestIDFromContext(r.Context())
tid := tenantID(r)
id := chi.URLParam(r, "id")

item, err := h.svc.Get(r.Context(), tid, id)
if err != nil {
h.log.Warn("failed to get example", zap.String("request_id", reqID), zap.String("id", id), zap.Error(err))
gokiterrors.WriteError(w, r, err)
return
}

response.Ok(w, r, item)
}

// Create handles POST /api/v1/examples
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
reqID := middleware.RequestIDFromContext(r.Context())
tid := tenantID(r)

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

// Update handles PUT /api/v1/examples/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
reqID := middleware.RequestIDFromContext(r.Context())
tid := tenantID(r)
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

// Delete handles DELETE /api/v1/examples/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
reqID := middleware.RequestIDFromContext(r.Context())
tid := tenantID(r)
id := chi.URLParam(r, "id")

if err := h.svc.Delete(r.Context(), tid, id); err != nil {
h.log.Error("failed to delete example", zap.String("request_id", reqID), zap.String("id", id), zap.Error(err))
gokiterrors.WriteError(w, r, err)
return
}

response.NoContent(w)
}
