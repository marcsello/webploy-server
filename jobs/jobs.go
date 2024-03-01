package jobs

import (
	"github.com/gdgvda/cron"
	"github.com/marcsello/webploy-server/site"
	"go.uber.org/zap"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
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

func wrapJob(logger *zap.Logger, job JobBase) func() {
	name := reflect.TypeOf(job).Elem().Name()
	j := &jobWrapper{
		name:         name,
		logger:       logger.With(zap.String("jobName", name)),
		runningMutex: &sync.Mutex{},
		execId:       atomic.Uint64{},
		job:          job,
	}
	return func() {
		j.Run()
	}
}

func InitJobRunner(logger *zap.Logger, sites site.Provider) (func() error, error) {

	c := cron.New(cron.WithSeconds())) // the builtin logger is super spammy and can only do info
	_, err := c.Schedule(cron.Every(time.Minute*1), wrapJob(logger, &janitorJob{sites}))
	if err != nil {
		return nil, err
	}

	runFn := func() error {
		c.Run()
		return nil
	}

	return runFn, nil
}
