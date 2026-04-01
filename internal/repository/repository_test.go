package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err = db.AutoMigrate(&Example{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return db
}

func TestPostgresRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgres(db)
	ctx := context.Background()

	example := &Example{ID: "id-1", Name: "Original", Description: "desc", TenantID: "tenant-1"}
	assert.NoError(t, repo.Create(ctx, example))

	created, err := repo.GetByID(ctx, "tenant-1", "id-1")
	assert.NoError(t, err)
	assert.Equal(t, "Original", created.Name)

	created.Name = "Updated"
	assert.NoError(t, repo.Update(ctx, created))

	updated, err := repo.GetByID(ctx, "tenant-1", "id-1")
	assert.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)

	assert.NoError(t, repo.Delete(ctx, "tenant-1", "id-1"))

	missing, err := repo.GetByID(ctx, "tenant-1", "id-1")
	assert.Error(t, err)
	assert.Nil(t, missing)
}

func TestPostgresRepo_ListPaginationAndTenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgres(db)
	ctx := context.Background()

	seed := []Example{
		{ID: "id-1", Name: "One", TenantID: "tenant-1"},
		{ID: "id-2", Name: "Two", TenantID: "tenant-1"},
		{ID: "id-3", Name: "Three", TenantID: "tenant-1"},
		{ID: "id-4", Name: "Other", TenantID: "tenant-2"},
	}
	for i := range seed {
		assert.NoError(t, repo.Create(ctx, &seed[i]))
	}

	items, total, err := repo.List(ctx, "tenant-1", ListParams{Page: 1, PageSize: 2})
	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, int64(3), total)

	items, total, err = repo.List(ctx, "tenant-2", ListParams{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "tenant-2", items[0].TenantID)
}

func TestPostgresRepo_ListSearchLengthGuard(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgres(db)

	longSearch := strings.Repeat("a", MaxSearchLength+1)
	items, total, err := repo.List(context.Background(), "tenant-1", ListParams{Page: 1, PageSize: 10, Search: longSearch})

	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, int64(0), total)
}
