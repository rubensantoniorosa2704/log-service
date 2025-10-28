package dto

import "github.com/google/uuid"

type CreateLogInput struct {
	ApplicationID uuid.UUID              `json:"application_id"`
	UserID        uuid.UUID              `json:"user_id"`
	Message       string                 `json:"message"`
	Level         string                 `json:"level"`
	Source        string                 `json:"source,omitempty"`
	Tags          map[string]string      `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type CreateLogOutput struct {
	ID            uuid.UUID              `json:"id"`
	ApplicationID uuid.UUID              `json:"application_id"`
	UserID        uuid.UUID              `json:"user_id"`
	Message       string                 `json:"message"`
	Level         string                 `json:"level"`
	Source        string                 `json:"source,omitempty"`
	Tags          map[string]string      `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     string                 `json:"timestamp"`
}

type LogOutput struct {
	ID            uuid.UUID              `json:"id"`
	ApplicationID uuid.UUID              `json:"application_id"`
	UserID        uuid.UUID              `json:"user_id"`
	Message       string                 `json:"message"`
	Level         string                 `json:"level"`
	Source        string                 `json:"source,omitempty"`
	Tags          map[string]string      `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     string                 `json:"timestamp"`
}
