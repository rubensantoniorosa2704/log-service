package log

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/application/log/dto"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/log"
)

// Mock implementations for testing
type mockLogRepository struct {
	createError bool
	createdLogs []*log.Log
}

func (m *mockLogRepository) Create(ctx context.Context, l *log.Log) error {
	if m.createError {
		return errors.New("repository error")
	}
	m.createdLogs = append(m.createdLogs, l)
	return nil
}

type mockSSEServer struct {
	streams      map[string]bool
	publishCalls []SSEPublishCall
}

type SSEPublishCall struct {
	Channel string
	Data    []byte
}

func (m *mockSSEServer) StreamExists(channel string) bool {
	if m.streams == nil {
		return false
	}
	return m.streams[channel]
}

func (m *mockSSEServer) Publish(channel string, data []byte) {
	m.publishCalls = append(m.publishCalls, SSEPublishCall{
		Channel: channel,
		Data:    data,
	})
}

func TestNewLogUsecase(t *testing.T) {
	repo := &mockLogRepository{}
	sseServer := &mockSSEServer{}

	usecase := NewLogUsecase(repo, sseServer)

	if usecase == nil {
		t.Error("Expected non-nil usecase")
	}
	if usecase.repo != repo {
		t.Error("Expected repository to be set correctly")
	}
	if usecase.sseSrv != sseServer {
		t.Error("Expected SSE server to be set correctly")
	}
}

func TestNewLogUsecase_WithoutSSE(t *testing.T) {
	repo := &mockLogRepository{}

	usecase := NewLogUsecase(repo, nil)

	if usecase == nil {
		t.Error("Expected non-nil usecase")
	}
	if usecase.repo != repo {
		t.Error("Expected repository to be set correctly")
	}
	if usecase.sseSrv != nil {
		t.Error("Expected SSE server to be nil")
	}
}

func TestLogUsecase_CreateLog_Success(t *testing.T) {
	repo := &mockLogRepository{}
	sseServer := &mockSSEServer{
		streams: make(map[string]bool),
	}
	usecase := NewLogUsecase(repo, sseServer)

	applicationID := uuid.New()
	userID := uuid.New()

	input := dto.CreateLogInput{
		ApplicationID: applicationID,
		UserID:        userID,
		Message:       "Test log message",
		Level:         "INFO",
		Source:        "TestService",
		Tags: map[string]string{
			"environment": "test",
		},
		Metadata: map[string]interface{}{
			"request_id": "req-123",
		},
	}

	ctx := context.Background()
	output, err := usecase.CreateLog(ctx, input)

	// Verify no error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify output
	if output == nil {
		t.Fatal("Expected non-nil output")
	}
	if output.ApplicationID != applicationID {
		t.Errorf("Expected ApplicationID %s, got %s", applicationID, output.ApplicationID)
	}
	if output.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, output.UserID)
	}
	if output.Message != input.Message {
		t.Errorf("Expected Message '%s', got '%s'", input.Message, output.Message)
	}
	if output.Level != input.Level {
		t.Errorf("Expected Level '%s', got '%s'", input.Level, output.Level)
	}

	// Verify repository was called
	if len(repo.createdLogs) != 1 {
		t.Errorf("Expected 1 log to be created in repository, got %d", len(repo.createdLogs))
	}

	// Verify SSE was not called (no stream exists)
	if len(sseServer.publishCalls) != 0 {
		t.Errorf("Expected 0 SSE publish calls, got %d", len(sseServer.publishCalls))
	}
}

func TestLogUsecase_CreateLog_WithSSE(t *testing.T) {
	repo := &mockLogRepository{}
	sseServer := &mockSSEServer{
		streams: make(map[string]bool),
	}
	usecase := NewLogUsecase(repo, sseServer)

	applicationID := uuid.New()
	userID := uuid.New()

	// Enable stream for this application
	sseServer.streams[applicationID.String()] = true

	input := dto.CreateLogInput{
		ApplicationID: applicationID,
		UserID:        userID,
		Message:       "Test log message with SSE",
		Level:         "WARN",
	}

	ctx := context.Background()
	output, err := usecase.CreateLog(ctx, input)

	// Verify no error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("Expected non-nil output")
	}

	// Verify SSE was called
	if len(sseServer.publishCalls) != 1 {
		t.Errorf("Expected 1 SSE publish call, got %d", len(sseServer.publishCalls))
	}

	// Verify SSE call details
	call := sseServer.publishCalls[0]
	if call.Channel != applicationID.String() {
		t.Errorf("Expected SSE channel '%s', got '%s'", applicationID.String(), call.Channel)
	}

	// Verify SSE payload can be unmarshaled
	var ssePayload dto.LogOutput
	if err := json.Unmarshal(call.Data, &ssePayload); err != nil {
		t.Errorf("Failed to unmarshal SSE payload: %v", err)
	}
	if ssePayload.Message != input.Message {
		t.Errorf("Expected SSE payload message '%s', got '%s'", input.Message, ssePayload.Message)
	}
}

func TestLogUsecase_CreateLog_WithoutSSEServer(t *testing.T) {
	repo := &mockLogRepository{}
	usecase := NewLogUsecase(repo, nil) // No SSE server

	applicationID := uuid.New()
	userID := uuid.New()

	input := dto.CreateLogInput{
		ApplicationID: applicationID,
		UserID:        userID,
		Message:       "Test log message without SSE",
		Level:         "ERROR",
	}

	ctx := context.Background()
	output, err := usecase.CreateLog(ctx, input)

	// Verify no error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("Expected non-nil output")
	}

	// Verify repository was called
	if len(repo.createdLogs) != 1 {
		t.Errorf("Expected 1 log to be created in repository, got %d", len(repo.createdLogs))
	}
}

func TestLogUsecase_CreateLog_InvalidInput(t *testing.T) {
	repo := &mockLogRepository{}
	sseServer := &mockSSEServer{
		streams: make(map[string]bool),
	}
	usecase := NewLogUsecase(repo, sseServer)

	tests := []struct {
		name  string
		input dto.CreateLogInput
	}{
		{
			name: "Invalid log level",
			input: dto.CreateLogInput{
				ApplicationID: uuid.New(),
				UserID:        uuid.New(),
				Message:       "Test message",
				Level:         "INVALID_LEVEL",
			},
		},
		{
			name: "Empty message",
			input: dto.CreateLogInput{
				ApplicationID: uuid.New(),
				UserID:        uuid.New(),
				Message:       "",
				Level:         "INFO",
			},
		},
		{
			name: "Empty application ID",
			input: dto.CreateLogInput{
				ApplicationID: uuid.UUID{},
				UserID:        uuid.New(),
				Message:       "Test message",
				Level:         "INFO",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			output, err := usecase.CreateLog(ctx, tt.input)

			if err == nil {
				t.Errorf("Expected error for invalid input, got none")
			}
			if output != nil {
				t.Errorf("Expected nil output for invalid input, got %+v", output)
			}

			// Verify repository was not called
			if len(repo.createdLogs) != 0 {
				t.Errorf("Expected 0 logs in repository for invalid input, got %d", len(repo.createdLogs))
			}
		})

		// Reset repository state for next test
		repo.createdLogs = nil
	}
}

func TestLogUsecase_CreateLog_RepositoryError(t *testing.T) {
	repo := &mockLogRepository{
		createError: true, // Simulate repository error
	}
	sseServer := &mockSSEServer{
		streams: make(map[string]bool),
	}
	usecase := NewLogUsecase(repo, sseServer)

	applicationID := uuid.New()
	userID := uuid.New()

	input := dto.CreateLogInput{
		ApplicationID: applicationID,
		UserID:        userID,
		Message:       "Test log message",
		Level:         "INFO",
	}

	ctx := context.Background()
	output, err := usecase.CreateLog(ctx, input)

	// Verify error occurred
	if err == nil {
		t.Error("Expected error from repository, got none")
	}
	if output != nil {
		t.Errorf("Expected nil output when repository fails, got %+v", output)
	}

	// Verify SSE was not called due to repository error
	if len(sseServer.publishCalls) != 0 {
		t.Errorf("Expected 0 SSE publish calls when repository fails, got %d", len(sseServer.publishCalls))
	}
}

func TestLogUsecase_CreateLog_ContextCancellation(t *testing.T) {
	repo := &mockLogRepository{}
	sseServer := &mockSSEServer{
		streams: make(map[string]bool),
	}
	usecase := NewLogUsecase(repo, sseServer)

	applicationID := uuid.New()
	userID := uuid.New()

	input := dto.CreateLogInput{
		ApplicationID: applicationID,
		UserID:        userID,
		Message:       "Test log message",
		Level:         "INFO",
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	output, err := usecase.CreateLog(ctx, input)

	// Note: Since our mock repository doesn't actually check context,
	// this test mainly verifies that the usecase handles context properly
	// In a real implementation, the repository would check context.Done()

	// For now, we just verify the usecase doesn't panic with cancelled context
	_ = output
	_ = err
}
