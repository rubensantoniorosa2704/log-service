package project

import (
	"errors"

	"github.com/google/uuid"
)

// Project represents a system, service, or application being monitored.
// This is the root of the 'Project' aggregate.
type Project struct {
	ID          uuid.UUID
	Name        string
	Description string
	Logs        []Log // A project can have many logs
}

func New(name, description string) (*Project, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	return &Project{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Logs:        []Log{},
	}, nil
}
