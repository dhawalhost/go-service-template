package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgxRepo is a high-performance read-optimised Repository implementation
// backed by pgx/v5. It is best suited for high-throughput List queries.
// For write operations, use the GORM-backed postgres implementation.
type pgxRepo struct{ pool *pgxpool.Pool }

// NewPgx creates a new pgx/v5-backed Repository implementation.
// This implementation is optimised for read-heavy List queries.
func NewPgx(pool *pgxpool.Pool) Repository { return &pgxRepo{pool: pool} }

func (r *pgxRepo) List(ctx context.Context, tenantID string, params ListParams) ([]Example, int64, error) {
	// Count query
	var total int64
	if params.Search != "" {
		err := r.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM examples WHERE tenant_id=$1 AND deleted_at IS NULL AND name ILIKE $2`,
			tenantID, "%"+params.Search+"%",
		).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("pgxRepo.List count: %w", err)
		}
	} else {
		err := r.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM examples WHERE tenant_id=$1 AND deleted_at IS NULL`,
			tenantID,
		).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("pgxRepo.List count: %w", err)
		}
	}

	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := 0
	if params.Page > 1 {
		offset = (params.Page - 1) * pageSize
	}

	// Data query
	var rows pgx.Rows
	var err error
	if params.Search != "" {
		rows, err = r.pool.Query(ctx,
			`SELECT id, name, description, tenant_id, created_at, updated_at
			 FROM examples
			 WHERE tenant_id=$1 AND deleted_at IS NULL AND name ILIKE $2
			 LIMIT $3 OFFSET $4`,
			tenantID, "%"+params.Search+"%", pageSize, offset,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, name, description, tenant_id, created_at, updated_at
			 FROM examples
			 WHERE tenant_id=$1 AND deleted_at IS NULL
			 LIMIT $2 OFFSET $3`,
			tenantID, pageSize, offset,
		)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("pgxRepo.List query: %w", err)
	}
	defer rows.Close()

	examples, err := pgx.CollectRows(rows, pgx.RowToStructByName[Example])
	if err != nil {
		return nil, 0, fmt.Errorf("pgxRepo.List collect: %w", err)
	}

	return examples, total, nil
}

// GetByID is not implemented in the pgx repository.
// Use the GORM postgres implementation for write operations.
func (r *pgxRepo) GetByID(_ context.Context, _, _ string) (*Example, error) {
	return nil, fmt.Errorf("not implemented: use postgres implementation for write operations")
}

// Create is not implemented in the pgx repository.
// Use the GORM postgres implementation for write operations.
func (r *pgxRepo) Create(_ context.Context, _ *Example) error {
	return fmt.Errorf("not implemented: use postgres implementation for write operations")
}

// Update is not implemented in the pgx repository.
// Use the GORM postgres implementation for write operations.
func (r *pgxRepo) Update(_ context.Context, _ *Example) error {
	return fmt.Errorf("not implemented: use postgres implementation for write operations")
}

// Delete is not implemented in the pgx repository.
// Use the GORM postgres implementation for write operations.
func (r *pgxRepo) Delete(_ context.Context, _, _ string) error {
	return fmt.Errorf("not implemented: use postgres implementation for write operations")
}

// ensure pgxRepo satisfies Repository at compile time.
var _ Repository = (*pgxRepo)(nil)
