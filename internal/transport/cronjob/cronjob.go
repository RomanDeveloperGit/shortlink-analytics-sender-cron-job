package cronjob

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type CronJob struct {
	zone   string
	spec   string
	cron   *cron.Cron
	logger *slog.Logger
}

type Options struct {
	Handler  func()
	Schedule string
	Zone     string
	Logger   *slog.Logger
}

func New(opts *Options) (*CronJob, error) {
	cronLogger := NewCronLogger(opts.Logger)

	location, err := time.LoadLocation(opts.Zone)
	if err != nil {
		panic(fmt.Errorf("failed to load location: %w", err))
	}

	c := cron.New(
		cron.WithLocation(location),
		cron.WithChain(
			cron.Recover(cronLogger),
			cron.SkipIfStillRunning(cronLogger),
		),
	)

	_, err = c.AddFunc(opts.Schedule, opts.Handler)
	if err != nil {
		panic(fmt.Errorf("failed to add job: %w", err))
	}

	return &CronJob{
		cron:   c,
		logger: opts.Logger,
		spec:   opts.Schedule,
		zone:   opts.Zone,
	}, nil
}

func (cj *CronJob) BackgroundRun() {
	cj.cron.Start()

	cj.logger.Info("cron job started",
		slog.String("schedule", cj.spec),
		slog.String("zone", cj.zone),
	)
}

func (cj *CronJob) Stop() {
	ctx := cj.cron.Stop()

	<-ctx.Done()
}
