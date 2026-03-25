package service

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/dhawalhost/gokit/cache"
	"github.com/dhawalhost/gokit/idgen"

	"github.com/dhawalhost/go-service-template/internal/repository"
)

// Service defines the business logic interface.
// RENAME_ME: replace Example with your domain entity.
type Service interface {
	List(ctx context.Context, tenantID string, params ListParams) ([]Example, int64, error)
	Get(ctx context.Context, tenantID, id string) (*Example, error)
	Create(ctx context.Context, tenantID string, req CreateRequest) (*Example, error)
	Update(ctx context.Context, tenantID, id string, req UpdateRequest) (*Example, error)
	Delete(ctx context.Context, tenantID, id string) error
}

type svc struct {
	repo  repository.Repository
	cache cache.Cache
	log   *zap.Logger
}

// New creates a new Service instance with the given dependencies.
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
	cacheKey := "example:" + id

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
	if cacheErr := s.cache.Delete(ctx, "example:"+id); cacheErr != nil {
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
	if cacheErr := s.cache.Delete(ctx, "example:"+id); cacheErr != nil {
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

// ensure svc satisfies Service at compile time.
var _ Service = (*svc)(nil)
