package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Reader defines read-only data access methods.
// RENAME_ME: rename Example to your entity name.
type Reader interface {
	// List returns paginated examples for a tenant with optional search filtering.
	// When tenantID is empty, no tenant filter is applied.
	// Supports ILIKE pattern matching on the name field.
	// Page and PageSize params control pagination; defaults to page 1, 20 items per page.
	// Returns the list of examples, total count (excluding limit), and any error.
	List(ctx context.Context, tenantID string, params ListParams) ([]Example, int64, error)
	// GetByID retrieves a single example by ID and tenant.
	// When tenantID is empty, no tenant filter is applied.
	// Returns NotFound error if not found or tenant mismatch.
	GetByID(ctx context.Context, tenantID, id string) (*Example, error)
}

// Writer defines write data access methods.
// RENAME_ME: rename Example to your entity name.
type Writer interface {
	// Create inserts a new example record.
	// The example should have all required fields set (ID, Name, TenantID).
	// Returns any database error (e.g., constraint violations).
	Create(ctx context.Context, example *Example) error
	// Update updates an existing example.
	// The example must match both ID and TenantID for the update to succeed.
	// Returns NotFound error if the example is not found.
	Update(ctx context.Context, example *Example) error
	// Delete soft-deletes an example by ID and tenant.
	// When tenantID is empty, no tenant filter is applied.
	// Soft deletes use the DeletedAt field and don't remove the record from the database.
	// Returns any database error.
	Delete(ctx context.Context, tenantID, id string) error
}

// Repository combines Reader and Writer into a single data access interface.
type Repository interface {
	Reader
	Writer
}

// ListParams holds pagination and filter parameters.
type ListParams struct {
	// Page is the page number (1-indexed). Defaults to 1.
	Page int
	// PageSize is the number of items per page. Defaults to 20.
	PageSize int
	// Search is a substring to search for in the name field (case-insensitive).
	Search string
}

// Example is the GORM/pgx model representing a resource.
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
