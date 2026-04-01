package service

import "time"

// Example is the service-layer domain entity.
// RENAME_ME: rename Example and all related types to your entity name.
type Example struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TenantID    string    `json:"tenant_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListParams holds pagination and filter parameters for listing examples.
type ListParams struct {
	Page     int
	PageSize int
	Search   string
}

// CreateRequest holds the data needed to create a new example.
type CreateRequest struct {
	Name        string
	Description string
}

// UpdateRequest holds the data needed to update an existing example.
type UpdateRequest struct {
	Name        string
	Description string
}
