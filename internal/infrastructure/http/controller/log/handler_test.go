package log

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/application/log/dto"
)

// Mock usecase for testing
type mockLogUsecase struct {
	createLogError  bool
	createLogOutput *dto.CreateLogOutput
}

func (m *mockLogUsecase) CreateLog(ctx context.Context, input dto.CreateLogInput) (*dto.CreateLogOutput, error) {
	if m.createLogError {
		return nil, errors.New("usecase error")
	}
	return m.createLogOutput, nil
}

func TestNewLogController(t *testing.T) {
	usecase := &mockLogUsecase{}
	controller := NewLogController(usecase)

	if controller == nil {
		t.Error("Expected non-nil controller")
	}
	if controller.Usecase != usecase {
		t.Error("Expected usecase to be set correctly")
	}
}

func TestLogController_CreateLogHandler_Success(t *testing.T) {
	// Setup mock usecase
	expectedOutput := &dto.CreateLogOutput{
		ID:            uuid.New(),
		ApplicationID: uuid.New(),
		UserID:        uuid.New(),
		Message:       "Test log message",
		Level:         "INFO",
		Source:        "TestService",
		Tags: map[string]string{
			"environment": "test",
		},
		Metadata: map[string]interface{}{
			"request_id": "req-123",
		},
		Timestamp: "2023-10-28T10:30:00Z",
	}

	usecase := &mockLogUsecase{
		createLogOutput: expectedOutput,
	}
	controller := NewLogController(usecase)

	// Create request body
	input := dto.CreateLogInput{
		ApplicationID: expectedOutput.ApplicationID,
		UserID:        expectedOutput.UserID,
		Message:       expectedOutput.Message,
		Level:         expectedOutput.Level,
		Source:        expectedOutput.Source,
		Tags:          expectedOutput.Tags,
		Metadata:      expectedOutput.Metadata,
	}

	jsonBody, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/v1/logs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	controller.CreateLogHandler(w, req)

	// Verify response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Verify content type
	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected content type '%s', got '%s'", expectedContentType, contentType)
	}

	// Verify response body
	var response dto.CreateLogOutput
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response.ID != expectedOutput.ID {
		t.Errorf("Expected ID %s, got %s", expectedOutput.ID, response.ID)
	}
	if response.Message != expectedOutput.Message {
		t.Errorf("Expected message '%s', got '%s'", expectedOutput.Message, response.Message)
	}
}

func TestLogController_CreateLogHandler_InvalidJSON(t *testing.T) {
	usecase := &mockLogUsecase{}
	controller := NewLogController(usecase)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/logs", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	controller.CreateLogHandler(w, req)

	// Verify response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Verify error message
	var errorResponse map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	expectedError := "Invalid request body format."
	if errorResponse["error"] != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, errorResponse["error"])
	}
}

func TestLogController_CreateLogHandler_UsecaseError(t *testing.T) {
	usecase := &mockLogUsecase{
		createLogError: true,
	}
	controller := NewLogController(usecase)

	// Create valid request body
	input := dto.CreateLogInput{
		ApplicationID: uuid.New(),
		UserID:        uuid.New(),
		Message:       "Test log message",
		Level:         "INFO",
	}

	jsonBody, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/api/v1/logs", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	controller.CreateLogHandler(w, req)

	// Verify response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	// Verify error message
	var errorResponse map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	expectedError := "An internal error occurred while creating the log."
	if errorResponse["error"] != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, errorResponse["error"])
	}
}

func TestLogController_CreateLogHandler_EmptyBody(t *testing.T) {
	usecase := &mockLogUsecase{}
	controller := NewLogController(usecase)

	// Create request with empty body
	req := httptest.NewRequest("POST", "/api/v1/logs", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	controller.CreateLogHandler(w, req)

	// Verify response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogController_CreateLogHandler_MissingContentType(t *testing.T) {
	usecase := &mockLogUsecase{}
	controller := NewLogController(usecase)

	// Create valid request body but without Content-Type header
	input := dto.CreateLogInput{
		ApplicationID: uuid.New(),
		UserID:        uuid.New(),
		Message:       "Test log message",
		Level:         "INFO",
	}

	jsonBody, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	// Create HTTP request without Content-Type header
	req := httptest.NewRequest("POST", "/api/v1/logs", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	// Setup mock to return success
	expectedOutput := &dto.CreateLogOutput{
		ID:            uuid.New(),
		ApplicationID: input.ApplicationID,
		UserID:        input.UserID,
		Message:       input.Message,
		Level:         input.Level,
		Timestamp:     "2023-10-28T10:30:00Z",
	}
	usecase.createLogOutput = expectedOutput

	// Execute handler
	controller.CreateLogHandler(w, req)

	// Should still work even without explicit Content-Type header
	// Go's json.Decoder can handle this
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestLogController_CreateLogHandler_DifferentLogLevels(t *testing.T) {
	levels := []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			// Setup mock usecase
			expectedOutput := &dto.CreateLogOutput{
				ID:            uuid.New(),
				ApplicationID: uuid.New(),
				UserID:        uuid.New(),
				Message:       "Test log message",
				Level:         level,
				Timestamp:     "2023-10-28T10:30:00Z",
			}

			usecase := &mockLogUsecase{
				createLogOutput: expectedOutput,
			}
			controller := NewLogController(usecase)

			// Create request body
			input := dto.CreateLogInput{
				ApplicationID: expectedOutput.ApplicationID,
				UserID:        expectedOutput.UserID,
				Message:       expectedOutput.Message,
				Level:         level,
			}

			jsonBody, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest("POST", "/api/v1/logs", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute handler
			controller.CreateLogHandler(w, req)

			// Verify response
			if w.Code != http.StatusCreated {
				t.Errorf("Expected status %d for level %s, got %d", http.StatusCreated, level, w.Code)
			}

			// Verify response body
			var response dto.CreateLogOutput
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to unmarshal response for level %s: %v", level, err)
			}

			if response.Level != level {
				t.Errorf("Expected level '%s', got '%s'", level, response.Level)
			}
		})
	}
}
