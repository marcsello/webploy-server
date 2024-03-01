package jobs

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/marcsello/webploy-server/site"
	"github.com/marcsello/webploy-server/utils"
	"go.uber.org/zap"
	"time"
)

// JobBase is Webploy's version of jobs, it is wrapped
type JobBase interface {
	Run(logger *zap.Logger)
}

type jobRunnerDaemon struct {
	scheduler gocron.Scheduler
}

func (jrd *jobRunnerDaemon) Start() error {
	jrd.scheduler.Start()
	return nil
}

func (jrd *jobRunnerDaemon) Destroy() error {
	return jrd.scheduler.Shutdown()
}

func (jrd *jobRunnerDaemon) ErrChan() <-chan error {
	return nil // seems like this is working lol
}

func InitJobRunner(logger *zap.Logger, sites site.Provider) (utils.Daemon, error) {

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	j := wrapJob(logger, &janitorJob{sites})
	var jobHandle gocron.Job
	jobHandle, err = scheduler.NewJob(gocron.DurationJob(1*time.Minute), gocron.NewTask(j.Run), gocron.WithSingletonMode(gocron.LimitModeReschedule))
	if err != nil {
		return nil, err
	}
	j.jobHandle = jobHandle

	return &jobRunnerDaemon{scheduler}, err
}
