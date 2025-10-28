package dto

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/log"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/valueobjects"
)

func TestToDomainLog(t *testing.T) {
	validApplicationID := uuid.New()
	validUserID := uuid.New()

	tests := []struct {
		name        string
		input       CreateLogInput
		expectError bool
		expectedErr string
	}{
		{
			name: "Valid input with minimal fields",
			input: CreateLogInput{
				ApplicationID: validApplicationID,
				UserID:        validUserID,
				Message:       "Test message",
				Level:         "INFO",
			},
			expectError: false,
		},
		{
			name: "Valid input with all fields",
			input: CreateLogInput{
				ApplicationID: validApplicationID,
				UserID:        validUserID,
				Message:       "Test message with metadata",
				Level:         "ERROR",
				Source:        "TestService",
				Tags: map[string]string{
					"environment": "test",
					"module":      "auth",
				},
				Metadata: map[string]interface{}{
					"request_id": "req-123",
					"duration":   150.5,
				},
			},
			expectError: false,
		},
		{
			name: "Invalid log level",
			input: CreateLogInput{
				ApplicationID: validApplicationID,
				UserID:        validUserID,
				Message:       "Test message",
				Level:         "INVALID_LEVEL",
			},
			expectError: true,
		},
		{
			name: "Empty message",
			input: CreateLogInput{
				ApplicationID: validApplicationID,
				UserID:        validUserID,
				Message:       "",
				Level:         "INFO",
			},
			expectError: true,
		},
		{
			name: "Empty application ID",
			input: CreateLogInput{
				ApplicationID: uuid.UUID{},
				UserID:        validUserID,
				Message:       "Test message",
				Level:         "INFO",
			},
			expectError: true,
		},
		{
			name: "Empty user ID",
			input: CreateLogInput{
				ApplicationID: validApplicationID,
				UserID:        uuid.UUID{},
				Message:       "Test message",
				Level:         "INFO",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToDomainLog(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				if result != nil {
					t.Errorf("Expected nil result when error occurs, got %+v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("Expected valid log, got nil")
				}

				// Verify basic fields
				if result.ApplicationID != tt.input.ApplicationID {
					t.Errorf("Expected ApplicationID %s, got %s", tt.input.ApplicationID, result.ApplicationID)
				}
				if result.UserID != tt.input.UserID {
					t.Errorf("Expected UserID %s, got %s", tt.input.UserID, result.UserID)
				}
				if result.Message != tt.input.Message {
					t.Errorf("Expected Message '%s', got '%s'", tt.input.Message, result.Message)
				}
				if string(result.Level) != tt.input.Level {
					t.Errorf("Expected Level '%s', got '%s'", tt.input.Level, result.Level)
				}

				// Verify optional fields
				if tt.input.Source != "" && result.Source != tt.input.Source {
					t.Errorf("Expected Source '%s', got '%s'", tt.input.Source, result.Source)
				}

				// Verify tags
				if tt.input.Tags != nil {
					for key, expectedValue := range tt.input.Tags {
						if actualValue, exists := result.Tags[key]; !exists {
							t.Errorf("Expected tag '%s' not found", key)
						} else if actualValue != expectedValue {
							t.Errorf("Expected tag '%s' = '%s', got '%s'", key, expectedValue, actualValue)
						}
					}
				}

				// Verify metadata
				if tt.input.Metadata != nil {
					for key, expectedValue := range tt.input.Metadata {
						if actualValue, exists := result.Metadata[key]; !exists {
							t.Errorf("Expected metadata '%s' not found", key)
						} else if actualValue != expectedValue {
							t.Errorf("Expected metadata '%s' = '%v', got '%v'", key, expectedValue, actualValue)
						}
					}
				}
			}
		})
	}
}

func TestLogToCreateLogOutput(t *testing.T) {
	// Create a domain log for testing
	applicationID := uuid.New()
	userID := uuid.New()
	message := "Test log message"
	level := valueobjects.LogLevelWarn
	source := "TestService"

	domainLog, err := log.New(message, level, applicationID, userID)
	if err != nil {
		t.Fatalf("Failed to create domain log: %v", err)
	}

	// Set optional fields
	domainLog.Source = source
	domainLog.Tags["environment"] = "test"
	domainLog.Tags["module"] = "auth"
	domainLog.Metadata["request_id"] = "req-123"
	domainLog.Metadata["duration"] = 150.5

	// Convert to output DTO
	output := LogToCreateLogOutput(domainLog)

	// Verify conversion
	if output.ID != domainLog.ID {
		t.Errorf("Expected ID %s, got %s", domainLog.ID, output.ID)
	}
	if output.ApplicationID != domainLog.ApplicationID {
		t.Errorf("Expected ApplicationID %s, got %s", domainLog.ApplicationID, output.ApplicationID)
	}
	if output.UserID != domainLog.UserID {
		t.Errorf("Expected UserID %s, got %s", domainLog.UserID, output.UserID)
	}
	if output.Message != domainLog.Message {
		t.Errorf("Expected Message '%s', got '%s'", domainLog.Message, output.Message)
	}
	if output.Level != string(domainLog.Level) {
		t.Errorf("Expected Level '%s', got '%s'", domainLog.Level, output.Level)
	}
	if output.Source != domainLog.Source {
		t.Errorf("Expected Source '%s', got '%s'", domainLog.Source, output.Source)
	}

	// Verify tags
	for key, expectedValue := range domainLog.Tags {
		if actualValue, exists := output.Tags[key]; !exists {
			t.Errorf("Expected tag '%s' not found in output", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected tag '%s' = '%s', got '%s'", key, expectedValue, actualValue)
		}
	}

	// Verify metadata
	for key, expectedValue := range domainLog.Metadata {
		if actualValue, exists := output.Metadata[key]; !exists {
			t.Errorf("Expected metadata '%s' not found in output", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected metadata '%s' = '%v', got '%v'", key, expectedValue, actualValue)
		}
	}

	// Verify timestamp format
	_, err = time.Parse(time.RFC3339, output.Timestamp)
	if err != nil {
		t.Errorf("Invalid timestamp format '%s': %v", output.Timestamp, err)
	}
}

func TestLogToLogOutput(t *testing.T) {
	// Create a domain log for testing
	applicationID := uuid.New()
	userID := uuid.New()
	message := "Test log message for LogOutput"
	level := valueobjects.LogLevelError

	domainLog, err := log.New(message, level, applicationID, userID)
	if err != nil {
		t.Fatalf("Failed to create domain log: %v", err)
	}

	// Set optional fields
	domainLog.Source = "LogOutputTestService"
	domainLog.Tags["test"] = "true"
	domainLog.Metadata["test_case"] = "LogToLogOutput"

	// Convert to LogOutput DTO
	output := LogToLogOutput(domainLog)

	// Verify conversion (similar to CreateLogOutput but for LogOutput type)
	if output.ID != domainLog.ID {
		t.Errorf("Expected ID %s, got %s", domainLog.ID, output.ID)
	}
	if output.ApplicationID != domainLog.ApplicationID {
		t.Errorf("Expected ApplicationID %s, got %s", domainLog.ApplicationID, output.ApplicationID)
	}
	if output.UserID != domainLog.UserID {
		t.Errorf("Expected UserID %s, got %s", domainLog.UserID, output.UserID)
	}
	if output.Message != domainLog.Message {
		t.Errorf("Expected Message '%s', got '%s'", domainLog.Message, output.Message)
	}
	if output.Level != string(domainLog.Level) {
		t.Errorf("Expected Level '%s', got '%s'", domainLog.Level, output.Level)
	}
	if output.Source != domainLog.Source {
		t.Errorf("Expected Source '%s', got '%s'", domainLog.Source, output.Source)
	}

	// Verify timestamp format
	_, err = time.Parse(time.RFC3339, output.Timestamp)
	if err != nil {
		t.Errorf("Invalid timestamp format '%s': %v", output.Timestamp, err)
	}
}

func TestToDomainLog_CaseInsensitiveLevel(t *testing.T) {
	validApplicationID := uuid.New()
	validUserID := uuid.New()

	testCases := []string{
		"info", "INFO", "Info", "InFo",
		"error", "ERROR", "Error", "ErRoR",
		"debug", "DEBUG", "Debug", "DeBuG",
	}

	for _, levelStr := range testCases {
		t.Run(levelStr, func(t *testing.T) {
			input := CreateLogInput{
				ApplicationID: validApplicationID,
				UserID:        validUserID,
				Message:       "Test message",
				Level:         levelStr,
			}

			result, err := ToDomainLog(input)
			if err != nil {
				t.Errorf("Unexpected error for level '%s': %v", levelStr, err)
			}
			if result == nil {
				t.Errorf("Expected valid log for level '%s', got nil", levelStr)
			}
		})
	}
}

func TestToDomainLog_EmptyOptionalFields(t *testing.T) {
	validApplicationID := uuid.New()
	validUserID := uuid.New()

	input := CreateLogInput{
		ApplicationID: validApplicationID,
		UserID:        validUserID,
		Message:       "Test message",
		Level:         "INFO",
		Source:        "",  // Empty source
		Tags:          nil, // Nil tags
		Metadata:      nil, // Nil metadata
	}

	result, err := ToDomainLog(input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify that empty/nil optional fields don't cause issues
	if result.Source != "" {
		t.Errorf("Expected empty source, got '%s'", result.Source)
	}

	// Maps should be initialized even if input was nil
	if result.Tags == nil {
		t.Errorf("Expected initialized Tags map, got nil")
	}
	if result.Metadata == nil {
		t.Errorf("Expected initialized Metadata map, got nil")
	}
}
