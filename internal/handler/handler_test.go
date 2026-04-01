package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	gokiterrors "github.com/dhawalhost/gokit/errors"

	handlermocks "github.com/dhawalhost/go-service-template/internal/handler/mocks"
	"github.com/dhawalhost/go-service-template/internal/service"
)

func withReqContext(r *http.Request) *http.Request {
	return r
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestHandler_List(t *testing.T) {
	svc := &handlermocks.Service{}
	svc.On("List", mock.Anything, "default", service.ListParams{Page: 1, PageSize: 20, Search: ""}).Return([]service.Example{}, int64(0), nil)
	h := New(svc, zaptest.NewLogger(t))
	r := withReqContext(httptest.NewRequest(http.MethodGet, "/api/v1/examples", nil))
	w := httptest.NewRecorder()

	h.List(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Get(t *testing.T) {
	svc := &handlermocks.Service{}
	svc.On("Get", mock.Anything, "default", "id-1").Return(&service.Example{ID: "id-1", Name: "test", TenantID: "default", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil)
	h := New(svc, zaptest.NewLogger(t))
	r := httptest.NewRequest(http.MethodGet, "/api/v1/examples/id-1", nil)
	r = withReqContext(withURLParam(r, "id", "id-1"))
	w := httptest.NewRecorder()

	h.Get(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Create_InvalidBody(t *testing.T) {
	svc := &handlermocks.Service{}
	h := New(svc, zaptest.NewLogger(t))
	r := withReqContext(httptest.NewRequest(http.MethodPost, "/api/v1/examples", bytes.NewBufferString("{")))
	w := httptest.NewRecorder()

	h.Create(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Create_Success(t *testing.T) {
	svc := &handlermocks.Service{}
	svc.On("Create", mock.Anything, "default", service.CreateRequest{Name: "new", Description: "desc"}).
		Return(&service.Example{ID: "id-1", Name: "new", Description: "desc", TenantID: "default", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil)

	h := New(svc, zaptest.NewLogger(t))
	r := withReqContext(httptest.NewRequest(http.MethodPost, "/api/v1/examples", bytes.NewBufferString(`{"name":"new","description":"desc"}`)))
	w := httptest.NewRecorder()

	h.Create(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Update_NotFound(t *testing.T) {
	svc := &handlermocks.Service{}
	svc.On("Update", mock.Anything, "default", "id-1", service.UpdateRequest{Name: "new", Description: "desc"}).
		Return(nil, gokiterrors.NotFound("EXAMPLE_NOT_FOUND", "missing"))

	h := New(svc, zaptest.NewLogger(t))
	r := httptest.NewRequest(http.MethodPut, "/api/v1/examples/id-1", bytes.NewBufferString(`{"name":"new","description":"desc"}`))
	r = withReqContext(withURLParam(r, "id", "id-1"))
	w := httptest.NewRecorder()

	h.Update(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Delete_Success(t *testing.T) {
	svc := &handlermocks.Service{}
	svc.On("Delete", mock.Anything, "default", "id-1").Return(nil)

	h := New(svc, zaptest.NewLogger(t))
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/examples/id-1", nil)
	r = withReqContext(withURLParam(r, "id", "id-1"))
	w := httptest.NewRecorder()

	h.Delete(w, r)

	assert.Equal(t, http.StatusNoContent, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_PatternAndRouter(t *testing.T) {
	h := New(&handlermocks.Service{}, zaptest.NewLogger(t))
	assert.Equal(t, "/api/v1/examples", h.Pattern())
	assert.NotNil(t, h.Router())
}
