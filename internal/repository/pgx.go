package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgxRepo is a read-optimised Reader implementation backed by pgx/v5.
// It is best suited for high-throughput List queries.
// For write operations, use the GORM-backed postgresRepo.
type pgxRepo struct{ pool *pgxpool.Pool }

// NewPgx returns a pgx/v5-backed Reader optimised for read-heavy workloads.
func NewPgx(pool *pgxpool.Pool) Reader { return &pgxRepo{pool: pool} }

func (r *pgxRepo) List(ctx context.Context, tenantID string, params ListParams) ([]Example, int64, error) {
	args := []any{tenantID}
	where := `WHERE tenant_id=$1 AND deleted_at IS NULL`
	if params.Search != "" {
		args = append(args, "%"+params.Search+"%")
		where += ` AND name ILIKE $2`
	}

	var total int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM examples `+where,
		args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("pgxRepo.List count: %w", err)
	}

	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := 0
	if params.Page > 1 {
		offset = (params.Page - 1) * pageSize
	}

	n := len(args)
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(
			`SELECT id, name, description, tenant_id, created_at, updated_at
			 FROM examples %s LIMIT $%d OFFSET $%d`,
			where, n+1, n+2,
		),
		append(args, pageSize, offset)...,
	)
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

func (r *pgxRepo) GetByID(ctx context.Context, tenantID, id string) (*Example, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, tenant_id, created_at, updated_at
		 FROM examples WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`,
		id, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("pgxRepo.GetByID query: %w", err)
	}
	defer rows.Close()

	ex, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[Example])
	if err != nil {
		return nil, fmt.Errorf("pgxRepo.GetByID: %w", err)
	}
	return &ex, nil
}

// ensure pgxRepo satisfies Reader at compile time.
var _ Reader = (*pgxRepo)(nil)
