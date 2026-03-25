package repository

import (
	"context"

	gokiterrors "github.com/dhawalhost/gokit/errors"
	"gorm.io/gorm"
)

type postgresRepo struct{ db *gorm.DB }

// NewPostgres creates a new GORM-backed Repository implementation.
func NewPostgres(db *gorm.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) List(ctx context.Context, tenantID string, params ListParams) ([]Example, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&Example{}).Where("tenant_id = ?", tenantID)

	if params.Search != "" {
		query = query.Where("name ILIKE ?", "%"+params.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := 0
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
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

	if err == gorm.ErrRecordNotFound {
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
	return r.db.WithContext(ctx).Save(example).Error
}

func (r *postgresRepo) Delete(ctx context.Context, tenantID, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&Example{}).Error
}

// ensure postgresRepo satisfies Repository at compile time.
var _ Repository = (*postgresRepo)(nil)
