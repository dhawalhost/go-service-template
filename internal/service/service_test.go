package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/dhawalhost/go-service-template/internal/repository"
	servicemocks "github.com/dhawalhost/go-service-template/internal/service/mocks"
)

func TestService_List(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*servicemocks.Repository, *servicemocks.Cache)
		wantTotal int64
		wantCount int
		wantErr   bool
	}{
		{
			name: "success",
			setup: func(repo *servicemocks.Repository, _ *servicemocks.Cache) {
				repo.On("List", mock.Anything, "tenant-1", repository.ListParams{Page: 1, PageSize: 10, Search: ""}).
					Return([]repository.Example{{ID: "id-1", Name: "name", TenantID: "tenant-1"}}, int64(1), nil)
			},
			wantTotal: 1,
			wantCount: 1,
		},
		{
			name: "repository error",
			setup: func(repo *servicemocks.Repository, _ *servicemocks.Cache) {
				repo.On("List", mock.Anything, "tenant-1", repository.ListParams{Page: 1, PageSize: 10, Search: ""}).
					Return(nil, int64(0), errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &servicemocks.Repository{}
			cache := &servicemocks.Cache{}
			tt.setup(repo, cache)

			s := New(repo, cache, zaptest.NewLogger(t))
			items, total, err := s.List(context.Background(), "tenant-1", ListParams{Page: 1, PageSize: 10})

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, items)
				assert.Equal(t, int64(0), total)
			} else {
				assert.NoError(t, err)
				assert.Len(t, items, tt.wantCount)
				assert.Equal(t, tt.wantTotal, total)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Get(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*servicemocks.Repository, *servicemocks.Cache)
		wantName string
		wantErr  bool
	}{
		{
			name: "cache hit",
			setup: func(_ *servicemocks.Repository, cache *servicemocks.Cache) {
				cached, _ := json.Marshal(Example{ID: "id-1", Name: "cached", TenantID: "tenant-1"})
				cache.On("Get", mock.Anything, "example:tenant-1:id-1").Return(string(cached), nil)
			},
			wantName: "cached",
		},
		{
			name: "cache miss then repository",
			setup: func(repo *servicemocks.Repository, cache *servicemocks.Cache) {
				cache.On("Get", mock.Anything, "example:tenant-1:id-1").Return("", errors.New("miss"))
				repo.On("GetByID", mock.Anything, "tenant-1", "id-1").Return(&repository.Example{ID: "id-1", Name: "db", TenantID: "tenant-1"}, nil)
				cache.On("Set", mock.Anything, "example:tenant-1:id-1", mock.Anything, 5*time.Minute).Return(nil)
			},
			wantName: "db",
		},
		{
			name: "repository error",
			setup: func(repo *servicemocks.Repository, cache *servicemocks.Cache) {
				cache.On("Get", mock.Anything, "example:tenant-1:id-1").Return("", errors.New("miss"))
				repo.On("GetByID", mock.Anything, "tenant-1", "id-1").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &servicemocks.Repository{}
			cache := &servicemocks.Cache{}
			tt.setup(repo, cache)

			s := New(repo, cache, zaptest.NewLogger(t))
			item, err := s.Get(context.Background(), "tenant-1", "id-1")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantName, item.Name)
			}

			repo.AssertExpectations(t)
			cache.AssertExpectations(t)
		})
	}
}

func TestService_Create(t *testing.T) {
	repo := &servicemocks.Repository{}
	cache := &servicemocks.Cache{}
	repo.On("Create", mock.Anything, mock.MatchedBy(func(e *repository.Example) bool {
		return e.Name == "new" && e.TenantID == "tenant-1" && e.ID != ""
	})).Return(nil)

	s := New(repo, cache, zaptest.NewLogger(t))
	item, err := s.Create(context.Background(), "tenant-1", CreateRequest{Name: "new", Description: "desc"})

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, "new", item.Name)
	assert.Equal(t, "tenant-1", item.TenantID)
	repo.AssertExpectations(t)
}

func TestService_Update(t *testing.T) {
	repo := &servicemocks.Repository{}
	cache := &servicemocks.Cache{}
	repo.On("GetByID", mock.Anything, "tenant-1", "id-1").Return(&repository.Example{ID: "id-1", Name: "old", TenantID: "tenant-1"}, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(e *repository.Example) bool {
		return e.ID == "id-1" && e.Name == "new-name"
	})).Return(nil)
	cache.On("Delete", mock.Anything, []string{"example:tenant-1:id-1"}).Return(nil)

	s := New(repo, cache, zaptest.NewLogger(t))
	item, err := s.Update(context.Background(), "tenant-1", "id-1", UpdateRequest{Name: "new-name", Description: "new-desc"})

	assert.NoError(t, err)
	assert.Equal(t, "new-name", item.Name)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestService_Delete(t *testing.T) {
	repo := &servicemocks.Repository{}
	cache := &servicemocks.Cache{}
	repo.On("Delete", mock.Anything, "tenant-1", "id-1").Return(nil)
	cache.On("Delete", mock.Anything, []string{"example:tenant-1:id-1"}).Return(errors.New("cache down"))

	s := New(repo, cache, zaptest.NewLogger(t))
	err := s.Delete(context.Background(), "tenant-1", "id-1")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestService_Get_SingleTenantCacheKey(t *testing.T) {
	repo := &servicemocks.Repository{}
	cache := &servicemocks.Cache{}
	cache.On("Get", mock.Anything, "example:id-1").Return("", errors.New("miss"))
	repo.On("GetByID", mock.Anything, "", "id-1").Return(&repository.Example{ID: "id-1", Name: "db"}, nil)
	cache.On("Set", mock.Anything, "example:id-1", mock.Anything, 5*time.Minute).Return(nil)

	s := New(repo, cache, zaptest.NewLogger(t))
	item, err := s.Get(context.Background(), "", "id-1")

	assert.NoError(t, err)
	assert.Equal(t, "db", item.Name)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}
