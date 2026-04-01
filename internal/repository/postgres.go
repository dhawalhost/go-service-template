package repository

import (
	"context"
	"errors"

	gokiterrors "github.com/dhawalhost/gokit/errors"
	"gorm.io/gorm"
)

const (
	// MaxPageSize is the maximum number of items allowed per page to prevent resource exhaustion
	MaxPageSize = 1000
	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 20
	// MaxSearchLength is the maximum allowed length for search queries
	MaxSearchLength = 100
)

type postgresRepo struct{ db *gorm.DB }

// NewPostgres creates a new GORM-backed Repository implementation.
func NewPostgres(db *gorm.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) List(ctx context.Context, tenantID string, params ListParams) ([]Example, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&Example{}).Where("tenant_id = ?", tenantID)

	// Validate and limit search query length to prevent expensive LIKE queries
	if params.Search != "" {
		if len(params.Search) > MaxSearchLength {
			return nil, 0, gokiterrors.BadRequest("SEARCH_TOO_LONG",
				"search query too long (max "+string(rune(MaxSearchLength))+" characters)")
		}
		query = query.Where("name ILIKE ?", "%"+params.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Validate and enforce maximum page size to prevent resource exhaustion
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset := 0
	if params.Page > 1 {
		offset = (params.Page - 1) * pageSize
	}

	var rows []Example
	if err := query.Limit(pageSize).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *postgresRepo) GetByID(ctx context.Context, tenantID, id string) (*Example, error) {
	var row Example
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&row).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gokiterrors.NotFound("EXAMPLE_NOT_FOUND", "example not found")
	}
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (r *postgresRepo) Create(ctx context.Context, example *Example) error {
	return r.db.WithContext(ctx).Create(example).Error
}

func (r *postgresRepo) Update(ctx context.Context, example *Example) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", example.ID, example.TenantID).
		Save(example)

	if result.Error != nil {
		return result.Error
	}

	// Verify the update actually affected a row
	if result.RowsAffected == 0 {
		return gokiterrors.NotFound("EXAMPLE_NOT_FOUND", "example not found or unauthorized")
	}

	return nil
}

func (r *postgresRepo) Delete(ctx context.Context, tenantID, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&Example{}).Error
}

// ensure postgresRepo satisfies Repository at compile time.
var _ Repository = (*postgresRepo)(nil)
