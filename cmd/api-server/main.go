package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hamba/logger/v2"
	"github.com/hamba/pkg/v2/http/server"
	"github.com/hamba/statter/v2"
	"github.com/huy125/financial-data-web/api"
	"github.com/huy125/financial-data-web/api/store"
)

func main() {
	var (
		apiKey string
		host string
		port string
		logLevel int
		logFormat string
	)

	flag.StringVar(&apiKey, "apiKey", "", "Alpha Vantage API Key, required for stocks endpoints")
	flag.StringVar(&host, "host", "localhost", "Server host, the default value is localhost")
	flag.StringVar(&port, "port", "8080", "Listen port for server, the default value is 8080")
	flag.IntVar(&logLevel, "log.level", int(logger.Debug), "Log level (debug = 5, info = 4, error = 2)")
	flag.StringVar(&logFormat, "log.format", "json", "Log format (json, text)")

	flag.Parse()

	log := logger.New(os.Stdout, logger.JSONFormat(), logger.Level(logLevel))
	cancel := log.WithTimestamp()
	defer cancel()

	stats := statter.New(statter.DiscardReporter, 10*time.Second)
	if apiKey == "" {
		log.Error("apiKey is required")
	}

	addr := net.JoinHostPort(host, port)
	store := store.NewInMemory()

	h := api.New(apiKey, store)
	server := server.GenericServer[context.Context]{
		Addr:    addr,
		Handler: h,
		Log:     log,
		Stats:   stats,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := server.Run(ctx); err != nil && err != http.ErrServerClosed {
		log.Error("Could not start server")
		cancel()
	}

	log.Info("Server terminated")
}
