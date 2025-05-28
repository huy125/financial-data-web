package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/huy125/financial-data-web/authenticator"
	"golang.org/x/oauth2"
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
	"github.com/huy125/financial-data-web/store"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func main() {
	authFlags := cmd.Flags{
		&cli.StringFlag{
			Name:     "auth0Domain",
			Usage:    "Auth0 domain",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "auth0ClientId",
			Usage:    "Auth0 client ID",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "auth0ClientSecret",
			Usage:    "Auth0 client secret",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "auth0CallbackUrl",
			Usage:    "Auth0 callback URL",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "hmacSecret",
			Usage:    "Secret key for HMAC operations in authentication",
			Required: true,
		},
	}

	flags := cmd.Flags{
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
		&cli.StringFlag{
			Name:  "dsn",
			Usage: "Data source name",
		},
		&cli.StringFlag{
			Name:  "algorithmPath",
			Usage: "Path to the file that contains scoring algorithm",
		},
	}.Merge(cmd.MonitoringFlags, authFlags)

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
	filePath := c.String("algorithmPath")
	host := c.String("host")
	port := c.String("port")
	dsn := c.String("dsn")

	auth0Domain := c.String("auth0Domain")
	auth0ClientId := c.String("auth0ClientId")
	auth0ClientSecret := c.String("auth0ClientSecret")
	auth0CallbackUrl := c.String("auth0CallbackUrl")
	hmacSecret := c.String("hmacSecret")

	if apiKey == "" {
		obsrv.Log.Error("apiKey is required")
		return nil
	}

	db, err := store.NewDB(store.WithDSN(dsn))
	if err != nil {
		obsrv.Log.Error("Could not set up store")
		return err
	}
	store := store.New(db)

	if auth0Domain == "" || auth0ClientId == "" || auth0ClientSecret == "" || auth0CallbackUrl == "" || hmacSecret == "" {
		obsrv.Log.Error("Authentication parameters are required")
		return nil
	}
	authProvider, err := oidc.NewProvider(
		context.Background(),
		"https://"+auth0Domain+"/",
	)
	if err != nil {
		obsrv.Log.Error("Could not set up authenticator provider")
		return err
	}

	authConfig := oauth2.Config{
		ClientID:     auth0ClientId,
		ClientSecret: auth0ClientSecret,
		RedirectURL:  auth0CallbackUrl,
		Endpoint:     authProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	auth, err := authenticator.New(authProvider, authConfig, []byte(hmacSecret))
	if err != nil {
		obsrv.Log.Error("Could not set up authenticator")
		return err
	}

	addr := net.JoinHostPort(host, port)
	h := api.New(apiKey, filePath, store, obsrv, auth)
	server := server.GenericServer[context.Context]{
		Addr:    addr,
		Handler: h,
		Log:     obsrv.Log,
		Stats:   obsrv.Stats,
	}

	if err := server.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		obsrv.Log.Error("Could not start server")
		cancel()
		return err
	}

	obsrv.Log.Info("Server terminated")

	return nil
}
