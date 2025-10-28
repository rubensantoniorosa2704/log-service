package dto

import (
	"time"

	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/log"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/valueobjects"
)

// ToDomainLog converts CreateLogInput DTO to domain log entity
func ToDomainLog(input CreateLogInput) (*log.Log, error) {
	level, err := valueobjects.NewLogLevel(input.Level)
	if err != nil {
		return nil, err
	}

	domainLog, err := log.New(input.Message, level, input.ApplicationID, input.UserID)
	if err != nil {
		return nil, err
	}

	// Set optional fields
	if input.Source != "" {
		domainLog.Source = input.Source
	}
	if input.Tags != nil {
		domainLog.Tags = input.Tags
	}
	if input.Metadata != nil {
		domainLog.Metadata = input.Metadata
	}

	return domainLog, nil
}

// LogToCreateLogOutput converts a domain log entity to CreateLogOutput DTO
func LogToCreateLogOutput(l *log.Log) CreateLogOutput {
	return CreateLogOutput{
		ID:            l.ID,
		ApplicationID: l.ApplicationID,
		UserID:        l.UserID,
		Message:       l.Message,
		Level:         string(l.Level),
		Source:        l.Source,
		Tags:          l.Tags,
		Metadata:      l.Metadata,
		Timestamp:     l.Timestamp.Format(time.RFC3339),
	}
}

// LogToLogOutput converts a domain log entity to LogOutput DTO
func LogToLogOutput(l *log.Log) LogOutput {
	return LogOutput{
		ID:            l.ID,
		ApplicationID: l.ApplicationID,
		UserID:        l.UserID,
		Message:       l.Message,
		Level:         string(l.Level),
		Source:        l.Source,
		Tags:          l.Tags,
		Metadata:      l.Metadata,
		Timestamp:     l.Timestamp.Format(time.RFC3339),
	}
}
