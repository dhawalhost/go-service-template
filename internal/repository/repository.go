package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Reader defines read-only data access methods.
// RENAME_ME: rename Example to your entity name.
type Reader interface {
	List(ctx context.Context, tenantID string, params ListParams) ([]Example, int64, error)
	GetByID(ctx context.Context, tenantID, id string) (*Example, error)
}

// Writer defines write data access methods.
// RENAME_ME: rename Example to your entity name.
type Writer interface {
	Create(ctx context.Context, example *Example) error
	Update(ctx context.Context, example *Example) error
	Delete(ctx context.Context, tenantID, id string) error
}

// Repository combines Reader and Writer into a single data access interface.
type Repository interface {
	Reader
	Writer
}

// ListParams holds pagination and filter parameters.
type ListParams struct {
	Page     int
	PageSize int
	Search   string
}

// Example is the GORM/pgx model.
// RENAME_ME: rename to your entity name.
type Example struct {
	ID          string         `gorm:"primaryKey"     db:"id"`
	Name        string         `gorm:"not null"       db:"name"`
	Description string         `                      db:"description"`
	TenantID    string         `gorm:"not null;index" db:"tenant_id"`
	CreatedAt   time.Time      `                      db:"created_at"`
	UpdatedAt   time.Time      `                      db:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"          db:"deleted_at"`
}
