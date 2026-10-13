package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RomanDeveloperGit/shortlink-analytics-sender-cron-job/internal/config"
	"github.com/RomanDeveloperGit/shortlink-analytics-sender-cron-job/internal/infra/grpcclient"
	"github.com/RomanDeveloperGit/shortlink-analytics-sender-cron-job/internal/infra/smtpclient"
	"github.com/RomanDeveloperGit/shortlink-analytics-sender-cron-job/internal/logger"
	"github.com/RomanDeveloperGit/shortlink-analytics-sender-cron-job/internal/service/analyticssender"
	"github.com/RomanDeveloperGit/shortlink-analytics-sender-cron-job/internal/transport/cronjob"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.MustLoad()
	log := logger.MustSetup(&logger.Options{
		File:           os.Stdout,
		Env:            cfg.Env,
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
	})

	grpcClient := grpcclient.NewClient(&grpcclient.Options{
		Host:              cfg.AnalyticsAPI.Host,
		Port:              cfg.AnalyticsAPI.Port,
		KeepAliveTime:     cfg.AnalyticsAPI.KeepAliveTime,
		ConnectionTimeout: cfg.AnalyticsAPI.ConnectTimeout,
		Logger:            log,
	})

	smtpClient := smtpclient.NewClient(&smtpclient.Options{
		Host:   cfg.SMTPServer.Host,
		Port:   cfg.SMTPServer.Port,
		Sender: cfg.SMTPServer.Sender,
	})

	analyticsSender := analyticssender.NewAnalyticsSender(log, smtpClient, grpcClient)

	cronHandler := func() {
		if err := analyticsSender.SendAllAnalytics(); err != nil {
			log.Error("failed to send analytics", slog.Any("error", err))
		}

		log.Info("analytics sent")
	}

	cronJob, err := cronjob.New(&cronjob.Options{
		Handler:  cronHandler,
		Schedule: cfg.Cron.Schedule,
		Zone:     cfg.Cron.Zone,
		Logger:   log,
	})

	if err != nil {
		log.Error("failed to create cron job", slog.Any("error", err))
		os.Exit(1)
	}

	cronJob.BackgroundRun()

	time.Sleep(1 * time.Millisecond)

	log.Info("server started")

	<-ctx.Done()

	log.Info("shutting down server")

	stop()

	cronJob.Stop()

	log.Info("server stopped")
}
