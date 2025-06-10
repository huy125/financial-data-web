package main

import (
	"context"
	"errors"
	"fmt"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/huy125/financial-data-web/authenticator"
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

// Config holds application configuration parameters
type Config struct {
	API struct {
		Key           string
		AlgorithmPath string
		Host          string
		Port          string
	}
	Auth struct {
		Domain       string
		ClientID     string
		ClientSecret string
		CallbackURL  string
		HmacSecret   string
		ApiAudience  string
	}
	DB struct {
		DSN string
	}
}

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
		&cli.StringFlag{
			Name:     "auth0ApiAudience",
			Usage:    "Auth0 audience for access token verification",
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

	cfg := loadConfig(c)

	// Validate API key
	if cfg.API.Key == "" {
		obsrv.Log.Error("apiKey is required")
		return errors.New("apiKey is required")
	}

	// Set up database
	store, err := setupDatabase(cfg)
	if err != nil {
		obsrv.Log.Error("Could not set up store", lctx.Error("error", err))
		return err
	}

	// Set up authenticator
	auth, err := setupAuthenticator(ctx, cfg, obsrv.Log)
	if err != nil {
		obsrv.Log.Error("Could not set up authenticator", lctx.Error("error", err))
		return err
	}

	addr := net.JoinHostPort(cfg.API.Host, cfg.API.Port)
	h := api.New(cfg.API.Key, cfg.API.AlgorithmPath, store, auth, obsrv)
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

// loadConfig loads configuration from CLI context
func loadConfig(c *cli.Context) Config {
	var cfg Config

	// API configuration
	cfg.API.Key = c.String("apiKey")
	cfg.API.AlgorithmPath = c.String("algorithmPath")
	cfg.API.Host = c.String("host")
	cfg.API.Port = c.String("port")

	// Auth configuration
	cfg.Auth.Domain = c.String("auth0Domain")
	cfg.Auth.ClientID = c.String("auth0ClientId")
	cfg.Auth.ClientSecret = c.String("auth0ClientSecret")
	cfg.Auth.CallbackURL = c.String("auth0CallbackUrl")
	cfg.Auth.HmacSecret = c.String("hmacSecret")
	cfg.Auth.ApiAudience = c.String("auth0ApiAudience")

	// DB configuration
	cfg.DB.DSN = c.String("dsn")

	return cfg
}

// setupDatabase creates and configures the database connection
func setupDatabase(cfg Config) (*store.Store, error) {
	db, err := store.NewDB(store.WithDSN(cfg.DB.DSN))
	if err != nil {
		return nil, fmt.Errorf("failed to set up database: %w", err)
	}

	return store.New(db), nil
}

// setupAuthenticator creates and configures an authenticator
func setupAuthenticator(ctx context.Context, cfg Config, log *logger.Logger) (*authenticator.Authenticator, error) {
	if cfg.Auth.Domain == "" {
		return nil, errors.New("domain is required")
	}
	if cfg.Auth.ClientID == "" {
		return nil, errors.New("client id is required")
	}
	if cfg.Auth.ClientSecret == "" {
		return nil, errors.New("client secret is required")
	}
	if cfg.Auth.CallbackURL == "" {
		return nil, errors.New("callback url is required")
	}
	if cfg.Auth.HmacSecret == "" {
		return nil, errors.New("hmac secret is required")
	}
	if cfg.Auth.ApiAudience == "" {
		return nil, errors.New("api audience is required")
	}

	auth, err := authenticator.New(
		ctx,
		cfg.Auth.Domain,
		authenticator.WithOAuthConfig(cfg.Auth.ClientID, cfg.Auth.ClientSecret, cfg.Auth.CallbackURL),
		authenticator.WithHmacSecret([]byte(cfg.Auth.HmacSecret)),
		authenticator.WithApiAudience(cfg.Auth.ApiAudience),
		authenticator.WithLogger(log),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize authenticator: %w", err)
	}

	return auth, nil
}
