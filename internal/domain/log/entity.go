package log

import (
	"time"

	"github.com/google/uuid"
	"github.com/rubensantoniorosa2704/LoggingSSE/internal/domain/valueobjects"
)

// Log represents a single log entry in the system.
// It is the main entity for logging events related to a project and user.
type Log struct {
	ID            uuid.UUID              `bson:"_id" json:"id"`
	Message       string                 `bson:"message" json:"message"`
	Level         valueobjects.LogLevel  `bson:"level" json:"level"`
	Timestamp     time.Time              `bson:"timestamp" json:"timestamp"`
	ApplicationID uuid.UUID              `bson:"application_id" json:"application_id"`
	UserID        uuid.UUID              `bson:"user_id" json:"user_id"`
	Source        string                 `bson:"source,omitempty" json:"source,omitempty"`     // Optional: source component/service
	Tags          map[string]string      `bson:"tags,omitempty" json:"tags,omitempty"`         // Optional: custom tags
	Metadata      map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"` // Optional: additional metadata
}

func New(message string, level valueobjects.LogLevel, applicationID, userID uuid.UUID) (*Log, error) {
	if message == "" {
		return nil, ErrMessageRequired
	}

	if !level.IsValid() {
		return nil, ErrLevelRequired
	}

	if applicationID == (uuid.UUID{}) {
		return nil, ErrApplicationIDInvalid
	}

	if userID == (uuid.UUID{}) {
		return nil, ErrUserIDInvalid
	}

	return &Log{
		ID:            uuid.New(),
		Message:       message,
		Level:         level,
		Timestamp:     time.Now(),
		ApplicationID: applicationID,
		UserID:        userID,
		Tags:          make(map[string]string),
		Metadata:      make(map[string]interface{}),
	}, nil
}
