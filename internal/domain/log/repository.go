package log

import (
	"context"
)

type LogRepository interface {
	Create(ctx context.Context, log *Log) error
}
