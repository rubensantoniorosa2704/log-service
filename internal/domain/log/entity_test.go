package log

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/valueobjects"
)

func TestLog_New(t *testing.T) {
	validApplicationID := uuid.New()
	validUserID := uuid.New()
	validLevel := valueobjects.LogLevelInfo
	validMessage := "Test log message"

	tests := []struct {
		name          string
		message       string
		level         valueobjects.LogLevel
		applicationID uuid.UUID
		userID        uuid.UUID
		expectError   bool
		expectedError error
	}{
		{
			name:          "Valid log creation",
			message:       validMessage,
			level:         validLevel,
			applicationID: validApplicationID,
			userID:        validUserID,
			expectError:   false,
		},
		{
			name:          "Empty message",
			message:       "",
			level:         validLevel,
			applicationID: validApplicationID,
			userID:        validUserID,
			expectError:   true,
			expectedError: ErrMessageRequired,
		},
		{
			name:          "Invalid log level",
			message:       validMessage,
			level:         valueobjects.LogLevel("INVALID"),
			applicationID: validApplicationID,
			userID:        validUserID,
			expectError:   true,
			expectedError: ErrLevelRequired,
		},
		{
			name:          "Empty application ID",
			message:       validMessage,
			level:         validLevel,
			applicationID: uuid.UUID{},
			userID:        validUserID,
			expectError:   true,
			expectedError: ErrApplicationIDInvalid,
		},
		{
			name:          "Empty user ID",
			message:       validMessage,
			level:         validLevel,
			applicationID: validApplicationID,
			userID:        uuid.UUID{},
			expectError:   true,
			expectedError: ErrUserIDInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := New(tt.message, tt.level, tt.applicationID, tt.userID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				if tt.expectedError != nil && err != tt.expectedError {
					t.Errorf("Expected error '%v', got '%v'", tt.expectedError, err)
				}
				if log != nil {
					t.Errorf("Expected nil log when error occurs, got %+v", log)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if log == nil {
					t.Errorf("Expected valid log, got nil")
				}

				// Verify log fields
				if log.Message != tt.message {
					t.Errorf("Expected message '%s', got '%s'", tt.message, log.Message)
				}
				if log.Level != tt.level {
					t.Errorf("Expected level '%s', got '%s'", tt.level, log.Level)
				}
				if log.ApplicationID != tt.applicationID {
					t.Errorf("Expected applicationID '%s', got '%s'", tt.applicationID, log.ApplicationID)
				}
				if log.UserID != tt.userID {
					t.Errorf("Expected userID '%s', got '%s'", tt.userID, log.UserID)
				}

				// Verify auto-generated fields
				if log.ID == (uuid.UUID{}) {
					t.Errorf("Expected generated ID, got empty UUID")
				}
				if log.Timestamp.IsZero() {
					t.Errorf("Expected generated timestamp, got zero time")
				}
				if time.Since(log.Timestamp) > time.Second {
					t.Errorf("Expected timestamp to be recent, got %v", log.Timestamp)
				}

				// Verify initialized maps
				if log.Tags == nil {
					t.Errorf("Expected initialized Tags map, got nil")
				}
				if log.Metadata == nil {
					t.Errorf("Expected initialized Metadata map, got nil")
				}
			}
		})
	}
}

func TestLog_NewWithOptionalFields(t *testing.T) {
	validApplicationID := uuid.New()
	validUserID := uuid.New()
	validLevel := valueobjects.LogLevelWarn
	validMessage := "Test log with optional fields"

	log, err := New(validMessage, validLevel, validApplicationID, validUserID)
	if err != nil {
		t.Fatalf("Unexpected error creating log: %v", err)
	}

	// Test setting optional fields
	log.Source = "TestService"
	log.Tags["environment"] = "test"
	log.Tags["module"] = "auth"
	log.Metadata["request_id"] = "req-123"
	log.Metadata["duration"] = 150.5

	// Verify optional fields
	if log.Source != "TestService" {
		t.Errorf("Expected source 'TestService', got '%s'", log.Source)
	}

	if log.Tags["environment"] != "test" {
		t.Errorf("Expected tag environment 'test', got '%s'", log.Tags["environment"])
	}

	if log.Tags["module"] != "auth" {
		t.Errorf("Expected tag module 'auth', got '%s'", log.Tags["module"])
	}

	if log.Metadata["request_id"] != "req-123" {
		t.Errorf("Expected metadata request_id 'req-123', got '%v'", log.Metadata["request_id"])
	}

	if log.Metadata["duration"] != 150.5 {
		t.Errorf("Expected metadata duration 150.5, got '%v'", log.Metadata["duration"])
	}
}

func TestLog_CreatedWithDifferentLevels(t *testing.T) {
	validApplicationID := uuid.New()
	validUserID := uuid.New()
	validMessage := "Test log message"

	levels := []valueobjects.LogLevel{
		valueobjects.LogLevelTrace,
		valueobjects.LogLevelDebug,
		valueobjects.LogLevelInfo,
		valueobjects.LogLevelWarn,
		valueobjects.LogLevelError,
		valueobjects.LogLevelFatal,
	}

	for _, level := range levels {
		t.Run(string(level), func(t *testing.T) {
			log, err := New(validMessage, level, validApplicationID, validUserID)
			if err != nil {
				t.Errorf("Unexpected error for level '%s': %v", level, err)
			}
			if log.Level != level {
				t.Errorf("Expected level '%s', got '%s'", level, log.Level)
			}
		})
	}
}

func TestLog_UniqueIDs(t *testing.T) {
	validApplicationID := uuid.New()
	validUserID := uuid.New()
	validLevel := valueobjects.LogLevelInfo
	validMessage := "Test log message"

	// Create multiple logs and verify they have unique IDs
	logs := make([]*Log, 10)
	for i := 0; i < 10; i++ {
		log, err := New(validMessage, validLevel, validApplicationID, validUserID)
		if err != nil {
			t.Fatalf("Unexpected error creating log %d: %v", i, err)
		}
		logs[i] = log
	}

	// Check that all IDs are unique
	idMap := make(map[uuid.UUID]bool)
	for i, log := range logs {
		if idMap[log.ID] {
			t.Errorf("Duplicate ID found at index %d: %s", i, log.ID)
		}
		idMap[log.ID] = true
	}
}

func TestLog_TimestampProgression(t *testing.T) {
	validApplicationID := uuid.New()
	validUserID := uuid.New()
	validLevel := valueobjects.LogLevelInfo
	validMessage := "Test log message"

	// Create two logs with a small delay
	log1, err := New(validMessage, validLevel, validApplicationID, validUserID)
	if err != nil {
		t.Fatalf("Unexpected error creating first log: %v", err)
	}

	time.Sleep(time.Millisecond * 10) // Small delay

	log2, err := New(validMessage, validLevel, validApplicationID, validUserID)
	if err != nil {
		t.Fatalf("Unexpected error creating second log: %v", err)
	}

	// Verify that second log has a later timestamp
	if !log2.Timestamp.After(log1.Timestamp) {
		t.Errorf("Expected log2 timestamp (%v) to be after log1 timestamp (%v)",
			log2.Timestamp, log1.Timestamp)
	}
}
