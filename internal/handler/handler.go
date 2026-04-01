package handler

import (
	"go.uber.org/zap"

	"github.com/dhawalhost/go-service-template/internal/service"
)

// Handler holds all HTTP handler dependencies for the examples resource.
// RENAME_ME: rename and extend for your service.
type Handler struct {
	svc service.Service
	log *zap.Logger
}

// New creates a new Handler instance with the given dependencies.
func New(svc service.Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}
