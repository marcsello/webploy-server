package deployment

import (
	"context"
	"io"
	"time"
	"webploy-server/deployment/info"
)

type Deployment interface {
	GetPath() string
	AddFile(ctx context.Context, relpath string, stream io.ReadCloser) error
	IsFinished() (bool, error)
	Finish() error
	Creator() (string, error)
	LastActivity() (time.Time, error)
	GetFullInfo() (info.DeploymentInfo, error)
}
