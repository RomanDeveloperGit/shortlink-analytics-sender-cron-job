package cronjob

import "log/slog"

type CronLogger struct {
	logger *slog.Logger
}

func NewCronLogger(logger *slog.Logger) *CronLogger {
	return &CronLogger{
		logger: logger,
	}
}

func (cl *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	cl.logger.Info(msg, keysAndValues...)
}

func (cl *CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	cl.logger.Error(msg, append(keysAndValues, "error", err)...)
}
