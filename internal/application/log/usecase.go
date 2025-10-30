package log

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rubensantoniorosa2704/LoggingSSE/internal/application/log/dto"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/log"
)

type LogUsecaseInterface interface {
	CreateLog(ctx context.Context, input dto.CreateLogInput) (*dto.CreateLogOutput, error)
}

// SSEPublisher interface for SSE server abstraction
type SSEPublisher interface {
	StreamExists(channel string) bool
	Publish(channel string, data []byte)
}

type LogUsecase struct {
	repo   log.LogRepository
	sseSrv SSEPublisher
}

// NewLogUsecase creates a new LogUsecase. Optionally pass an SSE server for real-time notifications.
func NewLogUsecase(repo log.LogRepository, sseSrv SSEPublisher) *LogUsecase {
	return &LogUsecase{repo: repo, sseSrv: sseSrv}
}

func (uc *LogUsecase) CreateLog(ctx context.Context, input dto.CreateLogInput) (*dto.CreateLogOutput, error) {
	// Convert DTO to domain entity
	newLog, err := dto.ToDomainLog(input)
	if err != nil {
		return nil, fmt.Errorf("invalid log data: %w", err)
	}

	// Save to repository
	if err := uc.repo.Create(ctx, newLog); err != nil {
		return nil, fmt.Errorf("failed to create log: %w", err)
	}

	// SSE notification: only if SSE server is present and there are clients for this application
	if uc.sseSrv != nil {
		channel := newLog.ApplicationID.String()
		if uc.sseSrv.StreamExists(channel) {
			payload, err := json.Marshal(dto.LogToLogOutput(newLog))
			if err != nil {
				// Log the error but don't fail the entire operation
				// The log was successfully saved to the database
				fmt.Printf("Warning: failed to marshal log for SSE: %v\n", err)
			} else {
				uc.sseSrv.Publish(channel, payload)
			}
		}
	}

	output := dto.LogToCreateLogOutput(newLog)
	return &output, nil
}
