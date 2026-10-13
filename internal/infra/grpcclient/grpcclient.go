package grpcclient

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	analyticsv1 "github.com/RomanDeveloperGit/shortlink-protos/gen/go/shortlink/analytics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
)

type Client struct {
	client                 *grpc.ClientConn
	AnalyticsServiceClient analyticsv1.AnalyticsServiceClient
	logger                 *slog.Logger
}

type Options struct {
	Host              string
	Port              int
	KeepAliveTime     time.Duration
	ConnectionTimeout time.Duration
	Logger            *slog.Logger
}

func NewClient(opts *Options) *Client {
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	conn, err := grpc.NewClient(
		addr,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    opts.KeepAliveTime,
			Timeout: opts.ConnectionTimeout,
		}),
		grpc.WithIdleTimeout(opts.ConnectionTimeout),
		grpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.ConnectionTimeout)
	defer cancel()

	conn.Connect()

	client := &Client{
		client:                 conn,
		logger:                 opts.Logger,
		AnalyticsServiceClient: analyticsv1.NewAnalyticsServiceClient(conn),
	}

	for {
		state := conn.GetState()

		if state == connectivity.Ready {
			return client
		}

		if !conn.WaitForStateChange(ctx, state) {
			panic("failed to connect to analytics api")
		}
	}
}

func (c *Client) Close() {
	err := c.client.Close()
	if err != nil {
		c.logger.Error("failed to close analytics api connection",
			slog.Any("error", err),
		)
	}
}
