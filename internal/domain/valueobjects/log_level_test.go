package valueobjects

import (
	"testing"
)

func TestNewLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    LogLevel
		expectError bool
	}{
		{
			name:        "Valid INFO level",
			input:       "INFO",
			expected:    LogLevelInfo,
			expectError: false,
		},
		{
			name:        "Valid ERROR level",
			input:       "ERROR",
			expected:    LogLevelError,
			expectError: false,
		},
		{
			name:        "Valid level with lowercase",
			input:       "debug",
			expected:    LogLevelDebug,
			expectError: false,
		},
		{
			name:        "Valid level with mixed case",
			input:       "WaRn",
			expected:    LogLevelWarn,
			expectError: false,
		},
		{
			name:        "Valid level with whitespace",
			input:       "  TRACE  ",
			expected:    LogLevelTrace,
			expectError: false,
		},
		{
			name:        "Invalid log level",
			input:       "INVALID",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Empty string",
			input:       "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewLogLevel(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Expected '%s', got '%s'", tt.expected, result)
				}
			}
		})
	}
}

func TestLogLevel_IsValid(t *testing.T) {
	validLevels := []LogLevel{
		LogLevelTrace,
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
		LogLevelFatal,
	}

	for _, level := range validLevels {
		t.Run(string(level), func(t *testing.T) {
			if !level.IsValid() {
				t.Errorf("Level '%s' should be valid", level)
			}
		})
	}

	invalidLevels := []LogLevel{
		"INVALID",
		"",
		"CUSTOM",
	}

	for _, level := range invalidLevels {
		t.Run(string(level), func(t *testing.T) {
			if level.IsValid() {
				t.Errorf("Level '%s' should be invalid", level)
			}
		})
	}
}

func TestLogLevel_Priority(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected int
	}{
		{LogLevelTrace, 1},
		{LogLevelDebug, 2},
		{LogLevelInfo, 3},
		{LogLevelWarn, 4},
		{LogLevelError, 5},
		{LogLevelFatal, 6},
		{"INVALID", 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			result := tt.level.Priority()
			if result != tt.expected {
				t.Errorf("Expected priority %d for level '%s', got %d", tt.expected, tt.level, result)
			}
		})
	}
}

func TestLogLevel_IsMoreSevereThan(t *testing.T) {
	tests := []struct {
		level1   LogLevel
		level2   LogLevel
		expected bool
	}{
		{LogLevelError, LogLevelWarn, true},
		{LogLevelFatal, LogLevelError, true},
		{LogLevelWarn, LogLevelInfo, true},
		{LogLevelInfo, LogLevelWarn, false},
		{LogLevelTrace, LogLevelDebug, false},
		{LogLevelInfo, LogLevelInfo, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.level1)+"_vs_"+string(tt.level2), func(t *testing.T) {
			result := tt.level1.IsMoreSevereThan(tt.level2)
			if result != tt.expected {
				t.Errorf("Expected %s.IsMoreSevereThan(%s) = %v, got %v",
					tt.level1, tt.level2, tt.expected, result)
			}
		})
	}
}

func TestLogLevel_IsLessSevereThan(t *testing.T) {
	tests := []struct {
		level1   LogLevel
		level2   LogLevel
		expected bool
	}{
		{LogLevelWarn, LogLevelError, true},
		{LogLevelInfo, LogLevelWarn, true},
		{LogLevelTrace, LogLevelFatal, true},
		{LogLevelError, LogLevelWarn, false},
		{LogLevelFatal, LogLevelError, false},
		{LogLevelInfo, LogLevelInfo, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.level1)+"_vs_"+string(tt.level2), func(t *testing.T) {
			result := tt.level1.IsLessSevereThan(tt.level2)
			if result != tt.expected {
				t.Errorf("Expected %s.IsLessSevereThan(%s) = %v, got %v",
					tt.level1, tt.level2, tt.expected, result)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelInfo, "INFO"},
		{LogLevelError, "ERROR"},
		{LogLevelDebug, "DEBUG"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("Expected String() = '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestValidLogLevels(t *testing.T) {
	levels := ValidLogLevels()
	expected := 6

	if len(levels) != expected {
		t.Errorf("Expected %d valid log levels, got %d", expected, len(levels))
	}

	// Verify all expected levels are present
	expectedLevels := map[LogLevel]bool{
		LogLevelTrace: false,
		LogLevelDebug: false,
		LogLevelInfo:  false,
		LogLevelWarn:  false,
		LogLevelError: false,
		LogLevelFatal: false,
	}

	for _, level := range levels {
		if _, exists := expectedLevels[level]; !exists {
			t.Errorf("Unexpected log level in ValidLogLevels(): %s", level)
		}
		expectedLevels[level] = true
	}

	// Verify all expected levels were found
	for level, found := range expectedLevels {
		if !found {
			t.Errorf("Expected log level '%s' not found in ValidLogLevels()", level)
		}
	}
}
