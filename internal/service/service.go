package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/dhawalhost/gokit/cache"
	"github.com/dhawalhost/gokit/idgen"

	"github.com/dhawalhost/go-service-template/internal/repository"
)

// Service defines the business logic interface for example operations.
// It implements the cache-aside pattern for reads and optional tenant isolation.
// RENAME_ME: replace Example with your domain entity.
type Service interface {
	// List returns paginated examples for a tenant with optional search filtering.
	// Pass an empty tenantID to operate in single-tenant mode.
	// Returns the list of examples, total count, and any error.
	List(ctx context.Context, tenantID string, params ListParams) ([]Example, int64, error)
	// Get retrieves a single example by ID for a given tenant.
	// Pass an empty tenantID to operate in single-tenant mode.
	// Uses cache-aside pattern: checks cache first, then database, then caches the result.
	// Returns NotFound error if the example doesn't exist or belongs to a different tenant.
	Get(ctx context.Context, tenantID, id string) (*Example, error)
	// Create creates a new example for a tenant.
	// Pass an empty tenantID to operate in single-tenant mode.
	// Returns the created example with auto-generated ID and timestamps, or any error.
	Create(ctx context.Context, tenantID string, req CreateRequest) (*Example, error)
	// Update updates an existing example for a tenant and invalidates its cache.
	// Pass an empty tenantID to operate in single-tenant mode.
	// Returns NotFound error if the example doesn't exist or belongs to a different tenant.
	Update(ctx context.Context, tenantID, id string, req UpdateRequest) (*Example, error)
	// Delete deletes an example for a tenant and invalidates its cache.
	// Pass an empty tenantID to operate in single-tenant mode.
	// Returns NotFound error if the example doesn't exist or belongs to a different tenant.
	Delete(ctx context.Context, tenantID, id string) error
}

type svc struct {
	repo  repository.Repository
	cache cache.Cache
	log   *zap.Logger
}

// New creates a new Service instance with the given dependencies.
// Dependencies:
//   - repo: data access layer for persistent storage
//   - cache: Redis cache for cache-aside pattern
//   - log: structured logger
func New(repo repository.Repository, cache cache.Cache, log *zap.Logger) Service {
	return &svc{repo: repo, cache: cache, log: log}
}

func (s *svc) List(ctx context.Context, tenantID string, params ListParams) ([]Example, int64, error) {
	rows, total, err := s.repo.List(ctx, tenantID, repository.ListParams{
		Page:     params.Page,
		PageSize: params.PageSize,
		Search:   params.Search,
	})
	if err != nil {
		return nil, 0, err
	}

	examples := make([]Example, 0, len(rows))
	for _, r := range rows {
		examples = append(examples, fromRepo(r))
	}
	return examples, total, nil
}

func (s *svc) Get(ctx context.Context, tenantID, id string) (*Example, error) {
	cacheKey := cacheKey(tenantID, id)

	// 1. Try cache first (cache-aside pattern)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var ex Example
		if jsonErr := json.Unmarshal([]byte(cached), &ex); jsonErr == nil {
			return &ex, nil
		}
	}

	// 2. Query database
	row, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	ex := fromRepo(*row)

	// 3. Marshal and cache with 5 minute TTL
	if data, jsonErr := json.Marshal(ex); jsonErr == nil {
		if cacheErr := s.cache.Set(ctx, cacheKey, string(data), 5*time.Minute); cacheErr != nil {
			s.log.Warn("failed to cache example", zap.String("id", id), zap.Error(cacheErr))
		}
	}

	return &ex, nil
}

func (s *svc) Create(ctx context.Context, tenantID string, req CreateRequest) (*Example, error) {
	row := &repository.Example{
		ID:          idgen.NewUUIDv7(),
		Name:        req.Name,
		Description: req.Description,
		TenantID:    tenantID,
	}

	if err := s.repo.Create(ctx, row); err != nil {
		return nil, err
	}

	ex := fromRepo(*row)
	return &ex, nil
}

func (s *svc) Update(ctx context.Context, tenantID, id string, req UpdateRequest) (*Example, error) {
	row, err := s.repo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	row.Name = req.Name
	row.Description = req.Description

	if err := s.repo.Update(ctx, row); err != nil {
		return nil, err
	}

	// Invalidate cache
	if cacheErr := s.cache.Delete(ctx, cacheKey(tenantID, id)); cacheErr != nil {
		s.log.Warn("failed to invalidate cache", zap.String("id", id), zap.Error(cacheErr))
	}

	ex := fromRepo(*row)
	return &ex, nil
}

func (s *svc) Delete(ctx context.Context, tenantID, id string) error {
	if err := s.repo.Delete(ctx, tenantID, id); err != nil {
		return err
	}

	// Invalidate cache
	if cacheErr := s.cache.Delete(ctx, cacheKey(tenantID, id)); cacheErr != nil {
		s.log.Warn("failed to invalidate cache", zap.String("id", id), zap.Error(cacheErr))
	}

	return nil
}

// fromRepo converts a repository.Example to a service.Example.
func fromRepo(r repository.Example) Example {
	return Example{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		TenantID:    r.TenantID,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func cacheKey(tenantID, id string) string {
	if strings.TrimSpace(tenantID) == "" {
		return "example:" + id
	}
	return "example:" + tenantID + ":" + id
}

// ensure svc satisfies Service at compile time.
var _ Service = (*svc)(nil)
