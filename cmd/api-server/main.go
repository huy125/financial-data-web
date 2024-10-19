package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
		port int
		env string
	)

	flag.StringVar(&apiKey, "apiKey", "", "Alpha Vantage API Key, required for stocks endpoints")
	flag.StringVar(&host, "host", "localhost", "Server host, the default value is localhost")
	flag.IntVar(&port, "port", 8080, "Listen port for server, the default value is 8080")
	flag.StringVar(&env, "env", "dev", "Environment where the server runs on, the default value is development")
	flag.Parse()

	minLvl := logger.Debug
	if env == "production" {
		minLvl = logger.Info
	}

	log := logger.New(os.Stdout, logger.JSONFormat(), minLvl)
	cancel := log.WithTimestamp()
	defer cancel()

	stats := statter.New(statter.DiscardReporter, 10*time.Second)
	if apiKey == "" {
		log.Error("apiKey is required")
	}

	addr := host + ":" + strconv.Itoa(port)
	h := api.New(apiKey, store)
	server := server.GenericServer[context.Context]{
		Addr: addr,
		Handler: h,
		Log: log,
		Stats: stats,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := server.Run(ctx); err != nil && err != http.ErrServerClosed {
		log.Error("Could not start server")
		cancel()
	}

	log.Info("Server terminated")
}
