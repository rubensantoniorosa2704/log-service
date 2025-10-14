package project

import (
	"context"

	"github.com/google/uuid"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type LogRepository interface {
	Create(ctx context.Context, log *Log) error
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*Log, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Log, error)
}
