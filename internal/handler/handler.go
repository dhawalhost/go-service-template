package handler

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/dhawalhost/go-service-template/internal/service"
	"github.com/dhawalhost/gokit/middleware"
)

// Handler holds all HTTP handler dependencies for the examples resource.
// RENAME_ME: rename and extend for your service.
type Handler struct {
	svc             service.Service
	log             *zap.Logger
	tenancyEnabled  bool
	defaultTenantID string
}

// New creates a new Handler instance with the given dependencies.
func New(svc service.Service, log *zap.Logger, tenancyEnabled bool, defaultTenantID string) *Handler {
	return &Handler{
		svc:             svc,
		log:             log,
		tenancyEnabled:  tenancyEnabled,
		defaultTenantID: strings.TrimSpace(defaultTenantID),
	}
}

func (h *Handler) tenantID(r *http.Request) string {
	if !h.tenancyEnabled {
		return ""
	}

	tid, ok := middleware.TenantIDFromContext(r.Context())
	if ok && strings.TrimSpace(tid) != "" {
		return tid
	}

	if h.defaultTenantID != "" {
		return h.defaultTenantID
	}

	return "default"
}
