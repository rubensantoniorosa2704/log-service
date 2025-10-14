package project

import (
	"time"

	"github.com/google/uuid"
)

// Log is a child entity within the Project aggregate.
// Represents a single log entry for a project.
type Log struct {
	ID        uuid.UUID
	Message   string
	Level     string
	Timestamp time.Time
	ProjectID uuid.UUID // Foreign key to Project
	UserID    uuid.UUID // ID of the user who generated the log (coming from the gateway)
}

func NewLog(message, level string, projectID, userID uuid.UUID) *Log {
	return &Log{
		ID:        uuid.New(),
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
		ProjectID: projectID,
		UserID:    userID,
	}
}
