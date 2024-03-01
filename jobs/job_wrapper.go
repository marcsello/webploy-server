package jobs

import (
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
	"reflect"
	"sync/atomic"
)

type jobWrapper struct {
	name       string
	logger     *zap.Logger
	execId     atomic.Uint64
	jobHandle  gocron.Job
	wrappedJob JobBase
}

func (w *jobWrapper) Run() {
	id := w.execId.Add(1)
	l := w.logger.With(zap.Uint64("execId", id))

	l.Debug("triggered")
	defer l.Debug("completed")
	w.wrappedJob.Run(l)
}

func wrapJob(logger *zap.Logger, job JobBase) *jobWrapper {
	name := reflect.TypeOf(job).Elem().Name()
	return &jobWrapper{
		name:       name,
		logger:     logger.With(zap.String("jobName", name)),
		execId:     atomic.Uint64{},
		wrappedJob: job,
	}
}
