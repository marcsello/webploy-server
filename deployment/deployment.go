package deployment

import (
	"context"
	"io"
	"time"
)

type Deployment interface {
	AddFile(ctx context.Context, relpath string, stream io.ReadCloser) error
	IsFinished() (bool, error)
	Finish() error
	Creator() (string, error)
	LastActivity() (time.Time, error)
}
