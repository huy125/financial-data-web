package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hamba/cmd/v2"
	"github.com/hamba/cmd/v2/observe"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/hamba/pkg/v2/http/server"
	"github.com/huy125/finscope/api"
	"github.com/huy125/finscope/pkg/authenticator"
	"github.com/huy125/finscope/store"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Config holds application configuration parameters.
type Config struct {
	API       APIConfig    `json:"api"`
	Pool      PoolConfig   `json:"pool"`
	Auth      AuthConfig   `json:"auth"`
	CookieCfg CookieConfig `json:"cookieConfig"`
}

// APIConfig holds API specific configurations.
type APIConfig struct {
	Key           string `json:"key"`
	AlgorithmPath string `json:"algorithmPath"`
	Host          string `json:"host"`
	Port          string `json:"port"`
}

// PoolConfig holds database specific configuration.
type PoolConfig struct {
	MaxConns        int32 `json:"maxConnections"`
	MinConns        int32 `json:"minConnections"`
	MaxConnIdleTime int32 `json:"maxConnectionIdleTime"`
	MaxConnLifetime int32 `json:"maxConnectionLifetime"`
}

// AuthConfig holds authenticator specific configurations.
type AuthConfig struct {
	ClientSecret string `json:"-"`
	HMACSecret   string `json:"-"`
	Domain       string `json:"domain"`
	ClientID     string `json:"clientId"`
	CallbackURL  string `json:"callbackUrl"`
	APIAudience  string `json:"apiAudience"`
	ClientOrigin string `json:"clientOrigin"`
}

// CookieConfig holds cookie specific configurations.
type CookieConfig struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	HttpOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

func main() {
	flags := cmd.Flags{
		&cli.StringFlag{
			Name:     "configPath",
			Usage:    "Path to API server configuration file",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "dsn",
			Usage: "Data source name",
		},
		&cli.StringFlag{
			Name:     "auth0ClientSecret",
			Usage:    "Auth0 client secret",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "hmacSecret",
			Usage:    "Secret key for HMAC operations in authentication",
			Required: true,
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

	obsrv, err := observe.NewFromCLI(c, "finscope-rest-server", &observe.Options{
		LogTimestamps: true,
		LogTimeFormat: logger.TimeFormatISO8601,
		StatsRuntime:  true,
		TracingAttrs:  []attribute.KeyValue{semconv.ServiceVersionKey.String("1.0.0")},
	})
	if err != nil {
		return err
	}
	defer obsrv.Close()

	dsn := c.String("dsn")
	if dsn == "" {
		obsrv.Log.Error("dsn is required")
	}

	cfg, err := parseCfg(c)
	if err != nil {
		obsrv.Log.Error("Could not prepare config", lctx.Error("error", err))
		return err
	}

	// Set up database
	store, err := setupDatabase(dsn, cfg.Pool)
	if err != nil {
		obsrv.Log.Error("Could not set up store", lctx.Error("error", err))
		return err
	}

	// Set up authenticator
	auth, err := setupAuthenticator(ctx, cfg.Auth, obsrv.Log)
	if err != nil {
		obsrv.Log.Error("Could not set up authenticator", lctx.Error("error", err))
		return err
	}

	cookieCfg := api.ServerCookieConfig{
		Name:     cfg.CookieCfg.Name,
		Path:     cfg.CookieCfg.Path,
		HttpOnly: cfg.CookieCfg.HttpOnly,
		Secure:   cfg.CookieCfg.Secure,
	}
	addr := net.JoinHostPort(cfg.API.Host, cfg.API.Port)
	h := api.New(cfg.API.Key, cfg.API.AlgorithmPath, cookieCfg, store, auth, obsrv)
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

// setupDatabase creates and configures the database connection.
func setupDatabase(dsn string, cfg PoolConfig) (*store.Store, error) {
	db, err := store.NewDB(
		store.WithDSN(dsn),
		store.WithMaxConns(cfg.MaxConns),
		store.WithMinConns(cfg.MinConns),
		store.WithMaxConnLifetime(time.Minute*time.Duration(cfg.MaxConnLifetime)),
		store.WithMaxConnIdleTime(time.Minute*time.Duration(cfg.MaxConnIdleTime)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set up database: %w", err)
	}

	return store.New(db), nil
}

// setupAuthenticator creates and configures an authenticator.
func setupAuthenticator(ctx context.Context, cfg AuthConfig, log *logger.Logger) (*authenticator.Authenticator, error) {
	auth, err := authenticator.New(
		ctx,
		cfg.Domain,
		authenticator.WithOAuthConfig(cfg.ClientID, cfg.ClientSecret, cfg.CallbackURL),
		authenticator.WithHMACSecret([]byte(cfg.HMACSecret)),
		authenticator.WithAPIAudience(cfg.APIAudience),
		authenticator.WithClientOrigin(cfg.ClientOrigin),
		authenticator.WithLogger(log),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize authenticator: %w", err)
	}

	return auth, nil
}

func parseCfg(c *cli.Context) (*Config, error) {
	cfgPath := c.String("configPath")
	cfg, err := loadConfigFromFile(cfgPath)
	if err != nil {
		return nil, err
	}

	// Override config with CLI parameters
	cfg.Auth.HMACSecret = c.String("hmacSecret")
	cfg.Auth.ClientSecret = c.String("auth0ClientSecret")

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadConfigFromFile loads configuration from file.
func loadConfigFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the Config object has all required fields filled in.
func (c *Config) Validate() error {
	if err := c.API.validate(); err != nil {
		return err
	}

	if err := c.Pool.validate(); err != nil {
		return err
	}

	if err := c.Auth.validate(); err != nil {
		return err
	}

	if err := c.CookieCfg.validate(); err != nil {
		return err
	}

	return nil
}

func (c *APIConfig) validate() error {
	if c.Key == "" {
		return errors.New("financial provider API key is required")
	}
	if c.AlgorithmPath == "" {
		return errors.New("scoring algorithm path is required")
	}
	if c.Host == "" {
		return errors.New("host is required")
	}
	if c.Port == "" {
		return errors.New("port is required")
	}

	return nil
}

func (c *PoolConfig) validate() error {
	if c.MaxConns <= 0 {
		return errors.New("max connections is required")
	}
	if c.MinConns <= 0 {
		return errors.New("min connections is required")
	}
	if c.MaxConnLifetime <= 0 {
		return errors.New("max connection lifetime is required")
	}
	if c.MaxConnIdleTime <= 0 {
		return errors.New("max connection idle time is required")
	}

	return nil
}

func (c *AuthConfig) validate() error {
	if c.Domain == "" {
		return errors.New("domain is required")
	}
	if c.ClientID == "" {
		return errors.New("client id is required")
	}
	if c.ClientSecret == "" {
		return errors.New("client secret is required")
	}
	if c.CallbackURL == "" {
		return errors.New("callback url is required")
	}
	if c.HMACSecret == "" {
		return errors.New("HMAC secret is required")
	}
	if c.APIAudience == "" {
		return errors.New("API audience is required")
	}
	if c.ClientOrigin == "" {
		return errors.New("client origin is required")
	}

	return nil
}

func (c *CookieConfig) validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Path == "" {
		return errors.New("path is required")
	}

	return nil
}
