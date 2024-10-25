package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hamba/cmd/v2"
	"github.com/hamba/cmd/v2/observe"
	"github.com/hamba/logger/v2"
	"github.com/hamba/pkg/v2/http/server"
	"github.com/huy125/financial-data-web/api"
	"github.com/huy125/financial-data-web/api/store"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func main() {
	var flags = cmd.Flags{
		&cli.StringFlag{
			Name:     "apiKey",
			Usage:    "AlphaVantage API Key, required for stocks endpoints",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "host",
			Usage: "Server host, default is localhost",
			Value: "localhost",
		},
		&cli.StringFlag{
			Name:  "port",
			Usage: "Listen port for server, default is 8080",
			Value: "8080",
		},
	}.Merge(cmd.MonitoringFlags)

	app := cli.NewApp()
	app.Name = "financial-server"
	app.Flags = flags
	app.Action = runServer

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Run CLI app
	if err := app.RunContext(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
	}
}

func runServer(c *cli.Context) error {
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	// lctx := convertToLoggerFields(c.StringSlice("log.ctx"))

	obsrv, err := observe.NewFromCLI(c, "financial-rest-server", &observe.Options{
		LogTimestamps: true,
		LogTimeFormat: logger.TimeFormatISO8601,
		StatsRuntime:  true,
		TracingAttrs:  []attribute.KeyValue{semconv.ServiceVersionKey.String("1.0.0")},
	})
	if err != nil {
		return err
	}
	defer obsrv.Close()

	apiKey := c.String("apiKey")
	host := c.String("host")
	port := c.String("port")

	if apiKey == "" {
		obsrv.Log.Error("apiKey is required")
		return nil
	}

	addr := net.JoinHostPort(host, port)
	store := store.NewInMemory()
	h := api.New(apiKey, store)
	server := server.GenericServer[context.Context]{
		Addr:    addr,
		Handler: h,
		Log:     obsrv.Log,
		Stats:   obsrv.Stats,
	}

	if err := server.Run(ctx); err != nil && err != http.ErrServerClosed {
		obsrv.Log.Error("Could not start server")
		cancel()
		return err
	}

	obsrv.Log.Info("Server terminated")

	return nil
}
