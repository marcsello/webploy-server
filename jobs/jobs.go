package jobs

import (
	"github.com/go-logr/zapr"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"webploy-server/site"
)

type JobBase interface {
	Run(logger *zap.Logger)
}

type jobWrapper struct {
	name         string
	logger       *zap.Logger
	runningMutex *sync.Mutex
	execId       atomic.Uint64
	job          JobBase
}

func (w *jobWrapper) Run() {
	id := w.execId.Add(1)
	l := w.logger.With(zap.Uint64("execId", id))

	locked := w.runningMutex.TryLock()
	if !locked {
		l.Warn("Skipping trigger for job, because it is already running")
		return
	}
	defer w.runningMutex.Unlock()

	l.Debug("triggered")
	defer l.Debug("completed")
	w.job.Run(l)
}

func wrapJob(logger *zap.Logger, job JobBase) *jobWrapper {
	name := reflect.TypeOf(job).Elem().Name()
	return &jobWrapper{
		name:         name,
		logger:       logger.With(zap.String("jobName", name)),
		runningMutex: &sync.Mutex{},
		execId:       atomic.Uint64{},
		job:          job,
	}
}

func InitJobRunner(logger *zap.Logger, sites site.Provider) (func() error, error) {

	c := cron.New(cron.WithSeconds(), cron.WithLogger(zapr.NewLogger(logger)))
	c.Schedule(cron.Every(time.Minute*1), wrapJob(logger, &janitorJob{sites}))

	runFn := func() error {
		c.Run()
		return nil
	}

	return runFn, nil
}
