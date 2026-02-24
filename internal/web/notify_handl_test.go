package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/valid"
)

type testNotifyService struct {
	createNotifyFunc      func(ctx context.Context, record *models.Record) error
	getNotifyStatByIDFunc func(ctx context.Context, id int64) (models.RecordStatus, error)
	deleteNotifyByIDFunc  func(ctx context.Context, id int64) error

	createCalls int
	getCalls    int
	deleteCalls int
}

func (s *testNotifyService) CreateNotify(ctx context.Context, record *models.Record) error {
	s.createCalls++
	if s.createNotifyFunc != nil {
		return s.createNotifyFunc(ctx, record)
	}
	return nil
}

func (s *testNotifyService) GetNotifyStatByID(ctx context.Context, id int64) (models.RecordStatus, error) {
	s.getCalls++
	if s.getNotifyStatByIDFunc != nil {
		return s.getNotifyStatByIDFunc(ctx, id)
	}
	return "", nil
}

func (s *testNotifyService) DeleteNotifyByID(ctx context.Context, id int64) error {
	s.deleteCalls++
	if s.deleteNotifyByIDFunc != nil {
		return s.deleteNotifyByIDFunc(ctx, id)
	}
	return nil
}

type testConfig struct {
	addr string
	path string
}

func (c testConfig) GetServerAddress() string   { return c.addr }
func (c testConfig) GetStaticFilesPath() string { return c.path }

var validatorSwapMu sync.Mutex

func withPermissiveWebValidator(t *testing.T) {
	t.Helper()

	validatorSwapMu.Lock()
	prev := valid.Validate

	v := validator.New()
	require.NoError(t, v.RegisterValidation("from_field", func(fl validator.FieldLevel) bool { return true }))
	require.NoError(t, v.RegisterValidation("to_field", func(fl validator.FieldLevel) bool { return true }))

	valid.Validate = v

	t.Cleanup(func() {
		valid.Validate = prev
		validatorSwapMu.Unlock()
	})
}

func requestWithURLParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)

	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func newCreateRequestBody(sendTime time.Time) string {
	return fmt.Sprintf(`{"message":"hello","dateTime":"%s","sendChan":"mail","from":"sender@example.com","to":["to@example.com"]}`,
		sendTime.UTC().Format(time.RFC3339),
	)
}

func TestServer_createNotify_BadJSON(t *testing.T) {
	service := &testNotifyService{}
	srv := &Server{service: service}

	req := httptest.NewRequest(http.MethodPost, "/notify", strings.NewReader(`{"message":`))
	rr := httptest.NewRecorder()

	srv.createNotify(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	assert.Contains(t, rr.Body.String(), "Bad Request")
	assert.Equal(t, 0, service.createCalls)
}

func TestServer_createNotify_ValidationError(t *testing.T) {
	service := &testNotifyService{}
	srv := &Server{service: service}

	body := `{"message":"hello","dateTime":"2000-01-01T00:00:00Z","sendChan":"mail","from":"bad","to":[]}`
	req := httptest.NewRequest(http.MethodPost, "/notify", strings.NewReader(body))
	rr := httptest.NewRecorder()

	srv.createNotify(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	assert.Contains(t, rr.Body.String(), "Validation error")
	assert.Equal(t, 0, service.createCalls)
}

func TestServer_createNotify_ServiceError(t *testing.T) {
	withPermissiveWebValidator(t)

	serviceErr := errors.New("service failed")
	service := &testNotifyService{
		createNotifyFunc: func(ctx context.Context, record *models.Record) error {
			require.NotNil(t, ctx)
			require.NotNil(t, record)
			return serviceErr
		},
	}
	srv := &Server{service: service}

	body := newCreateRequestBody(time.Now().Add(2 * time.Hour))
	req := httptest.NewRequest(http.MethodPost, "/notify", strings.NewReader(body))
	rr := httptest.NewRecorder()

	srv.createNotify(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
	assert.Contains(t, rr.Body.String(), "couldn`t create Notify")
	assert.Equal(t, 1, service.createCalls)
}

func TestServer_createNotify_Success(t *testing.T) {
	withPermissiveWebValidator(t)

	sendTime := time.Now().Add(3 * time.Hour).UTC().Truncate(time.Second)
	service := &testNotifyService{
		createNotifyFunc: func(ctx context.Context, record *models.Record) error {
			require.NotNil(t, ctx)
			require.NotNil(t, record)
			record.Id = 123
			assert.Equal(t, []byte("hello"), record.Data)
			assert.True(t, record.SendTime.Equal(sendTime))
			assert.Equal(t, models.RecordStatusWaiting, record.RecStat)
			assert.Equal(t, models.SendChanMail, record.SendChan)
			assert.Equal(t, "sender@example.com", record.From)
			assert.Equal(t, []string{"to@example.com"}, record.To)
			return nil
		},
	}
	srv := &Server{service: service}

	body := newCreateRequestBody(sendTime)
	req := httptest.NewRequest(http.MethodPost, "/notify", strings.NewReader(body))
	rr := httptest.NewRecorder()

	srv.createNotify(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	var resp PostCreateNotifyResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, int64(123), resp.ID)
	assert.Equal(t, "hello", resp.Message)
	assert.True(t, resp.DateTime.Equal(sendTime))
	assert.Equal(t, models.RecordStatusWaiting, resp.Status)
	assert.Equal(t, models.SendChanMail, resp.SendChan)
	assert.Equal(t, []string{"to@example.com"}, resp.To)
	assert.Equal(t, 1, service.createCalls)
}

func TestServer_createNotify_UsesDefaultFromForMail(t *testing.T) {
	withPermissiveWebValidator(t)

	sendTime := time.Now().Add(3 * time.Hour).UTC()
	service := &testNotifyService{
		createNotifyFunc: func(ctx context.Context, record *models.Record) error {
			require.NotNil(t, record)
			assert.Equal(t, "smtp-user@example.com", record.From)
			return nil
		},
	}
	srv := &Server{service: service, defaultFrom: "smtp-user@example.com"}

	body := fmt.Sprintf(`{"message":"hello","dateTime":"%s","sendChan":"mail","to":["to@example.com"]}`,
		sendTime.Format(time.RFC3339),
	)
	req := httptest.NewRequest(http.MethodPost, "/notify", strings.NewReader(body))
	rr := httptest.NewRecorder()

	srv.createNotify(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Result().StatusCode)
	assert.Equal(t, 1, service.createCalls)
}

func TestServer_getNotifyStatByID_InvalidID(t *testing.T) {
	service := &testNotifyService{}
	srv := &Server{service: service}

	req := httptest.NewRequest(http.MethodGet, "/notify/abc/", nil)
	req = requestWithURLParam(req, "id", "abc")
	rr := httptest.NewRecorder()

	srv.getNotifyStatByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	assert.Contains(t, rr.Body.String(), "invalid id format")
	assert.Equal(t, 0, service.getCalls)
}

func TestServer_getNotifyStatByID_ServiceError(t *testing.T) {
	serviceErr := errors.New("service failed")
	service := &testNotifyService{
		getNotifyStatByIDFunc: func(ctx context.Context, id int64) (models.RecordStatus, error) {
			require.NotNil(t, ctx)
			assert.Equal(t, int64(42), id)
			return "", serviceErr
		},
	}
	srv := &Server{service: service}

	req := httptest.NewRequest(http.MethodGet, "/notify/42/", nil)
	req = requestWithURLParam(req, "id", "42")
	rr := httptest.NewRecorder()

	srv.getNotifyStatByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
	assert.Contains(t, rr.Body.String(), "couldn`t get notify stat by ID")
	assert.Equal(t, 1, service.getCalls)
}

func TestServer_getNotifyStatByID_Success(t *testing.T) {
	service := &testNotifyService{
		getNotifyStatByIDFunc: func(ctx context.Context, id int64) (models.RecordStatus, error) {
			require.NotNil(t, ctx)
			assert.Equal(t, int64(42), id)
			return models.RecordStatusWaiting, nil
		},
	}
	srv := &Server{service: service}

	req := httptest.NewRequest(http.MethodGet, "/notify/42/", nil)
	req = requestWithURLParam(req, "id", "42")
	rr := httptest.NewRecorder()

	srv.getNotifyStatByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	assert.Equal(t, string(models.RecordStatusWaiting), rr.Body.String())
	assert.Equal(t, 1, service.getCalls)
}

func TestServer_deleteNotifyByID_InvalidID(t *testing.T) {
	service := &testNotifyService{}
	srv := &Server{service: service}

	req := httptest.NewRequest(http.MethodDelete, "/notify/abc/", nil)
	req = requestWithURLParam(req, "id", "abc")
	rr := httptest.NewRecorder()

	srv.deleteNotifyByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	assert.Contains(t, rr.Body.String(), "invalid id format")
	assert.Equal(t, 0, service.deleteCalls)
}

func TestServer_deleteNotifyByID_ServiceError(t *testing.T) {
	serviceErr := errors.New("service failed")
	service := &testNotifyService{
		deleteNotifyByIDFunc: func(ctx context.Context, id int64) error {
			require.NotNil(t, ctx)
			assert.Equal(t, int64(7), id)
			return serviceErr
		},
	}
	srv := &Server{service: service}

	req := httptest.NewRequest(http.MethodDelete, "/notify/7/", nil)
	req = requestWithURLParam(req, "id", "7")
	rr := httptest.NewRecorder()

	srv.deleteNotifyByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Result().StatusCode)
	assert.Contains(t, rr.Body.String(), "couldn`t delete notify by ID")
	assert.Equal(t, 1, service.deleteCalls)
}

func TestServer_deleteNotifyByID_Success(t *testing.T) {
	service := &testNotifyService{
		deleteNotifyByIDFunc: func(ctx context.Context, id int64) error {
			require.NotNil(t, ctx)
			assert.Equal(t, int64(7), id)
			return nil
		},
	}
	srv := &Server{service: service}

	req := httptest.NewRequest(http.MethodDelete, "/notify/7/", nil)
	req = requestWithURLParam(req, "id", "7")
	rr := httptest.NewRecorder()

	srv.deleteNotifyByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)
	assert.Empty(t, rr.Body.String())
	assert.Equal(t, 1, service.deleteCalls)
}

func TestNewServer_InitializesFields(t *testing.T) {
	cfg := testConfig{addr: ":8080", path: "internal/web/static"}
	service := &testNotifyService{}

	srv := NewServer(cfg, service)

	require.NotNil(t, srv)
	assert.Equal(t, ":8080", srv.adress)
	assert.Equal(t, "internal/web/static", srv.staticPath)
	assert.Same(t, service, srv.service)
	require.NotNil(t, srv.router)
	require.NotNil(t, srv.httpServer)
	assert.Equal(t, ":8080", srv.httpServer.Addr)
	assert.Same(t, srv.router, srv.httpServer.Handler)
}

func TestServer_routes_DispatchesAPIRoute(t *testing.T) {
	cfg := testConfig{addr: ":0", path: "internal/web/static"}
	service := &testNotifyService{}

	srv := NewServer(cfg, service)
	srv.routs()

	req := httptest.NewRequest(http.MethodGet, "/notify/not-a-number/", nil)
	rr := httptest.NewRecorder()

	srv.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Result().StatusCode)
	assert.Contains(t, rr.Body.String(), "invalid id format")
}
