package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dglazkoff/go-metrics/internal/logger"
	"github.com/dglazkoff/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) Get(ctx context.Context, name string) (models.Metrics, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockMetricsService) GetAll(ctx context.Context) ([]models.Metrics, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockMetricsService) Update(ctx context.Context, metric models.Metrics) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockMetricsService) UpdateList(ctx context.Context, metric []models.Metrics) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockMetricsService) PingDB(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestPingDB(t *testing.T) {
	err := logger.Initialize()
	assert.NoError(t, err)

	t.Run("success", func(tt *testing.T) {
		mockService := new(MockMetricsService)
		mockService.On("PingDB", mock.Anything).Return(nil)

		api := API{metricsService: mockService}
		handler := api.PingDB()

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(tt, http.StatusOK, rec.Result().StatusCode, "Expected HTTP status 200 OK")
		mockService.AssertCalled(tt, "PingDB", mock.Anything)
	})

	t.Run("db ping error", func(tt *testing.T) {
		mockService := new(MockMetricsService)
		mockService.On("PingDB", mock.Anything).Return(errors.New("mock DB ping error"))

		api := API{metricsService: mockService}
		handler := api.PingDB()

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(tt, http.StatusInternalServerError, rec.Result().StatusCode, "Expected HTTP status 500 Internal Server Error")
		mockService.AssertCalled(tt, "PingDB", mock.Anything)
	})
}
