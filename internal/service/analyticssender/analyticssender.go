package analyticssender

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	analyticsv1 "github.com/RomanDeveloperGit/shortlink-protos/gen/go/shortlink/analytics/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/RomanDeveloperGit/shortlink-analytics-sender-cron-job/internal/infra/grpcclient"
	"github.com/RomanDeveloperGit/shortlink-analytics-sender-cron-job/internal/infra/smtpclient"
)

type AnalyticsSender struct {
	smtpClient *smtpclient.Client
	grpcClient *grpcclient.Client
	logger     *slog.Logger
}

func NewAnalyticsSender(
	logger *slog.Logger,
	smtpClient *smtpclient.Client,
	grpcClient *grpcclient.Client,
) *AnalyticsSender {
	return &AnalyticsSender{
		logger:     logger,
		smtpClient: smtpClient,
		grpcClient: grpcClient,
	}
}

func (as *AnalyticsSender) SendAllAnalytics() error {
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	data, err := as.grpcClient.AnalyticsServiceClient.GetStatsByAuthorsInPeriod(
		context.Background(),
		&analyticsv1.GetStatsByAuthorsInPeriodRequest{
			From: timestamppb.New(from),
			To:   timestamppb.New(to),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get analytics: %w", err)
	}

	// Если прям очень по-хорошему делать, то воркер пул надо воткнуть, чтобы не заспамить планировщик своими горутинами, если стата большая
	// Ну и, конечно, тут должно это всё выполняться батчами, а не так, как я сразу всё всем одним разом. В реальности так не получится при высокой нагрузке

	wg := sync.WaitGroup{}

	wg.Add(len(data.GetStatsByAuthor()))
	for authorEmail, stats := range data.GetStatsByAuthor() {
		go func() {
			defer wg.Done()

			message := strings.Builder{}

			message.WriteString(fmt.Sprintf(
				"Добрый день!\n\nОтправляем вам статистику за период c %s до %s по всем вашим ссылкам:\n",
				from.Format(time.Stamp),
				to.Format(time.Stamp),
			))

			for i, stat := range stats.GetItems() {
				message.WriteString(fmt.Sprintf(
					"%d. Ссылка на %s с shortCode \"%s\" была использована %d раз",
					i+1,
					stat.GetFullUrl(),
					stat.GetShortLink(),
					stat.GetTotalVisits(),
				))

				if stat.GetFirstVisitAt() != nil {
					message.WriteString(fmt.Sprintf(
						", первый раз %s, последний раз %s\n",
						stat.GetFirstVisitAt().AsTime().Format(time.Stamp),
						stat.GetLastVisitAt().AsTime().Format(time.Stamp),
					))
				} else {
					message.WriteString("\n")
				}
			}

			err := as.smtpClient.Send([]string{authorEmail}, message.String())
			if err != nil {
				as.logger.Error("failed to send analytics",
					slog.String("email", authorEmail),
					slog.Any("error", err),
				)
			}
		}()
	}

	wg.Wait()

	return nil
}
