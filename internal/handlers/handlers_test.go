package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OlegRozh/subscriptions-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, sub *models.Subscription) (int64, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) Get(ctx context.Context, id int64) (*models.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscription), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, sub *models.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetSum(ctx context.Context, userID, serviceName, start, end string) (int, error) {
	args := m.Called(ctx, userID, serviceName, start, end)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockRepository) GetList(ctx context.Context, userID, serviceName string) ([]models.Subscription, error) {
	args := m.Called(ctx, userID, serviceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Subscription), args.Error(1)
}

func (m *MockRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func setupTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Tests

func TestCreate(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		mockReturn     int64
		mockError      error
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "success",
			body: models.Subscription{
				ServiceName: "Yandex Plus",
				Price:       399,
				UserId:      "550e8400-e29b-41d4-a716-446655440000",
				StartMonth:  time.Now(),
			},
			mockReturn:     1,
			expectedStatus: http.StatusCreated,
			expectedBody:   map[string]interface{}{"id": float64(1)},
		},
		{
			name:           "invalid JSON",
			body:           "{invalid json}",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "Invalid request body"},
		},
		{
			name: "validation failed - empty service name",
			body: models.Subscription{
				ServiceName: "",
				Price:       100,
				UserId:      "123",
				StartMonth:  time.Now(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "Invalid request body"},
		},
		{
			name: "validation failed - zero price",
			body: models.Subscription{
				ServiceName: "Test",
				Price:       0,
				UserId:      "123",
				StartMonth:  time.Now(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "Invalid request body"},
		},
		{
			name: "validation failed - empty user id",
			body: models.Subscription{
				ServiceName: "Test",
				Price:       100,
				UserId:      "",
				StartMonth:  time.Now(),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "Invalid request body"},
		},
		{
			name: "repository error",
			body: models.Subscription{
				ServiceName: "Test",
				Price:       100,
				UserId:      "123",
				StartMonth:  time.Now(),
			},
			mockReturn:     0,
			mockError:      errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "Failed to create subscription"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			if tt.mockReturn != 0 || tt.mockError != nil {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Subscription")).Return(tt.mockReturn, tt.mockError)
			}
			handler := NewHandler(mockRepo, setupTestLogger())

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/subscriptions", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Create(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.Equal(t, tt.expectedBody, resp)

			if tt.mockReturn != 0 || tt.mockError != nil {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "Create")
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockReturn     *models.Subscription
		mockError      error
		expectedStatus int
		expectedBody   any
	}{
		{
			name: "success",
			id:   "1",
			mockReturn: &models.Subscription{
				Id:          1,
				ServiceName: "Yandex Plus",
				Price:       399,
				UserId:      "123",
			},
			expectedStatus: http.StatusOK,
			expectedBody: &models.Subscription{
				Id:          1,
				ServiceName: "Yandex Plus",
				Price:       399,
				UserId:      "123",
			},
		},
		{
			name:           "not found",
			id:             "999",
			mockError:      errors.New("not found"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]any{"error": "Subscription not found"},
		},
		{
			name:           "invalid id",
			id:             "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]any{"error": "Invalid subscription id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			if tt.mockReturn != nil || tt.mockError != nil {
				mockRepo.On("Get", mock.Anything, mock.Anything).Return(tt.mockReturn, tt.mockError)
			}
			handler := NewHandler(mockRepo, setupTestLogger())

			req := httptest.NewRequest("GET", "/subscriptions/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()

			handler.Get(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp models.Subscription
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.mockReturn.Id, resp.Id)
				assert.Equal(t, tt.mockReturn.ServiceName, resp.ServiceName)
			} else {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.expectedBody, resp)
			}

			if tt.mockReturn != nil || tt.mockError != nil {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestGetSum(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockReturn     int
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "success with only user_id",
			queryParams:    "user_id=123",
			mockReturn:     1500,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"total_sum": float64(1500)},
		},
		{
			name:           "success with all filters",
			queryParams:    "user_id=123&service_name=Yandex&start=2025-01-01&end=2025-12-31",
			mockReturn:     2500,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"total_sum": float64(2500)},
		},
		{
			name:           "missing user_id",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "user_id is required"},
		},
		{
			name:           "repository error",
			queryParams:    "user_id=123",
			mockError:      errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{"error": "Internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := new(MockRepository)
			if tt.mockReturn != 0 || tt.mockError != nil {
				mockRepo.On("GetSum", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.mockReturn, tt.mockError)
			}
			handler := NewHandler(mockRepo, setupTestLogger())

			req := httptest.NewRequest("GET", "/subscriptions/sum?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.GetSum(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			assert.Equal(t, tt.expectedBody, resp)

			if tt.mockReturn != 0 || tt.mockError != nil {
				mockRepo.AssertExpectations(t)
			} else {
				mockRepo.AssertNotCalled(t, "GetSum")
			}
		})
	}
}

func TestGetList(t *testing.T) {
	mockRepo := new(MockRepository)
	handler := NewHandler(mockRepo, setupTestLogger())

	expected := []models.Subscription{
		{Id: 1, ServiceName: "Yandex Plus", Price: 399, UserId: "123"},
		{Id: 2, ServiceName: "Netflix", Price: 799, UserId: "123"},
	}

	mockRepo.On("GetList", mock.Anything, "123", "").Return(expected, nil)

	req := httptest.NewRequest("GET", "/subscriptions?user_id=123", nil)
	w := httptest.NewRecorder()

	handler.GetList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string][]models.Subscription
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp["subscriptions"], 2)
	mockRepo.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockError      error
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "success",
			id:             "1",
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
			expectedBody:   nil,
		},
		{
			name:           "not found",
			id:             "999",
			mockError:      errors.New("not found"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   map[string]interface{}{"error": "Subscription not found"},
		},
		{
			name:           "invalid id",
			id:             "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{"error": "Invalid subscription id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			if tt.mockError != nil {
				mockRepo.On("Delete", mock.Anything, mock.Anything).Return(tt.mockError)
			} else if tt.id != "invalid" {
				mockRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
			}
			handler := NewHandler(mockRepo, setupTestLogger())
			req := httptest.NewRequest("DELETE", "/subscriptions/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()

			handler.Delete(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, tt.expectedBody, resp)
			}

			if tt.id != "invalid" {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}
