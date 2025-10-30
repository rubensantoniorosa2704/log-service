package valueobjects

import (
	"fmt"
	"strings"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelTrace LogLevel = "TRACE"
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

var validLogLevelsMap = map[LogLevel]bool{
	LogLevelTrace: true,
	LogLevelDebug: true,
	LogLevelInfo:  true,
	LogLevelWarn:  true,
	LogLevelError: true,
	LogLevelFatal: true,
}

// ValidLogLevels returns all valid log levels
func ValidLogLevels() []LogLevel {
	return []LogLevel{
		LogLevelTrace,
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
		LogLevelFatal,
	}
}

// NewLogLevel creates a new LogLevel with validation
func NewLogLevel(level string) (LogLevel, error) {
	normalized := LogLevel(strings.ToUpper(strings.TrimSpace(level)))

	if !normalized.IsValid() {
		return "", fmt.Errorf("invalid log level '%s', valid levels are: %v", level, ValidLogLevels())
	}

	return normalized, nil
}

// IsValid checks if the log level is valid using O(1) map lookup
func (l LogLevel) IsValid() bool {
	return validLogLevelsMap[l]
}

// String returns the string representation of the log level
func (l LogLevel) String() string {
	return string(l)
}

// Priority returns the numeric priority of the log level (higher number = higher priority)
func (l LogLevel) Priority() int {
	switch l {
	case LogLevelTrace:
		return 1
	case LogLevelDebug:
		return 2
	case LogLevelInfo:
		return 3
	case LogLevelWarn:
		return 4
	case LogLevelError:
		return 5
	case LogLevelFatal:
		return 6
	default:
		return 0
	}
}

// IsMoreSevereThan checks if this log level is more severe than another
func (l LogLevel) IsMoreSevereThan(other LogLevel) bool {
	return l.Priority() > other.Priority()
}

// IsLessSevereThan checks if this log level is less severe than another
func (l LogLevel) IsLessSevereThan(other LogLevel) bool {
	return l.Priority() < other.Priority()
}
