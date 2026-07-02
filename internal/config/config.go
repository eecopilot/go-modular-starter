package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type LookupFunc func(string) (string, bool)

type Config struct {
	App     AppConfig
	HTTP    HTTPConfig
	Log     LogConfig
	Userkit UserkitConfig
}

type AppConfig struct {
	Name      string
	Env       string
	Version   string
	Commit    string
	BuildTime string
}

type HTTPConfig struct {
	Addr               string
	ReadHeaderTimeout  time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
	MaxHeaderBytes     int
	BodyLimitBytes     int64
	CORSAllowedOrigins []string
}

type LogConfig struct {
	Level  string
	Format string
}

type UserkitConfig struct {
	Enabled           bool
	DatabaseURL       string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration
	JWTSecret         string
	JWTIssuer         string
	JWTAudience       string
	TokenTTL          time.Duration
	BCryptCost        int
}

func Load() (Config, error) {
	return LoadFromLookup(os.LookupEnv)
}

func LoadFromLookup(lookup LookupFunc) (Config, error) {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	cfg := Config{
		App: AppConfig{
			Name:      envString(lookup, "APP_NAME", "go-modular-starter"),
			Env:       envString(lookup, "APP_ENV", "development"),
			Version:   envString(lookup, "APP_VERSION", "dev"),
			Commit:    envString(lookup, "APP_COMMIT", "unknown"),
			BuildTime: envString(lookup, "APP_BUILD_TIME", "unknown"),
		},
		HTTP: HTTPConfig{
			Addr:               envString(lookup, "HTTP_ADDR", ":8080"),
			ReadHeaderTimeout:  envDuration(lookup, "HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
			ReadTimeout:        envDuration(lookup, "HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:       envDuration(lookup, "HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:        envDuration(lookup, "HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:    envDuration(lookup, "HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
			MaxHeaderBytes:     envInt(lookup, "HTTP_MAX_HEADER_BYTES", 1<<20),
			BodyLimitBytes:     int64(envInt(lookup, "HTTP_BODY_LIMIT_BYTES", 1<<20)),
			CORSAllowedOrigins: envCSV(lookup, "HTTP_CORS_ALLOWED_ORIGINS"),
		},
		Log: LogConfig{
			Level:  strings.ToLower(envString(lookup, "LOG_LEVEL", "info")),
			Format: strings.ToLower(envString(lookup, "LOG_FORMAT", "text")),
		},
		Userkit: UserkitConfig{
			Enabled:           envBool(lookup, "USERKIT_ENABLED", false),
			DatabaseURL:       envString(lookup, "USERKIT_DATABASE_URL", ""),
			DBMaxOpenConns:    envInt(lookup, "USERKIT_DB_MAX_OPEN_CONNS", 10),
			DBMaxIdleConns:    envInt(lookup, "USERKIT_DB_MAX_IDLE_CONNS", 5),
			DBConnMaxLifetime: envDuration(lookup, "USERKIT_DB_CONN_MAX_LIFETIME", 30*time.Minute),
			DBConnMaxIdleTime: envDuration(lookup, "USERKIT_DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
			JWTSecret:         envString(lookup, "USERKIT_JWT_SECRET", ""),
			JWTIssuer:         envString(lookup, "USERKIT_JWT_ISSUER", "go-modular-starter"),
			JWTAudience:       envString(lookup, "USERKIT_JWT_AUDIENCE", "go-modular-starter-users"),
			TokenTTL:          envDuration(lookup, "USERKIT_TOKEN_TTL", 24*time.Hour),
			BCryptCost:        envInt(lookup, "USERKIT_BCRYPT_COST", 0),
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	var errs []error

	if strings.TrimSpace(c.App.Name) == "" {
		errs = append(errs, errors.New("APP_NAME is required"))
	}
	if strings.TrimSpace(c.HTTP.Addr) == "" {
		errs = append(errs, errors.New("HTTP_ADDR is required"))
	}
	if c.HTTP.ReadHeaderTimeout <= 0 {
		errs = append(errs, errors.New("HTTP_READ_HEADER_TIMEOUT must be positive"))
	}
	if c.HTTP.ReadTimeout <= 0 {
		errs = append(errs, errors.New("HTTP_READ_TIMEOUT must be positive"))
	}
	if c.HTTP.WriteTimeout <= 0 {
		errs = append(errs, errors.New("HTTP_WRITE_TIMEOUT must be positive"))
	}
	if c.HTTP.IdleTimeout <= 0 {
		errs = append(errs, errors.New("HTTP_IDLE_TIMEOUT must be positive"))
	}
	if c.HTTP.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("HTTP_SHUTDOWN_TIMEOUT must be positive"))
	}
	if c.HTTP.MaxHeaderBytes <= 0 {
		errs = append(errs, errors.New("HTTP_MAX_HEADER_BYTES must be positive"))
	}
	if c.HTTP.BodyLimitBytes <= 0 {
		errs = append(errs, errors.New("HTTP_BODY_LIMIT_BYTES must be positive"))
	}
	if !validLogLevel(c.Log.Level) {
		errs = append(errs, fmt.Errorf("LOG_LEVEL %q is invalid", c.Log.Level))
	}
	if c.Log.Format != "text" && c.Log.Format != "json" {
		errs = append(errs, fmt.Errorf("LOG_FORMAT %q is invalid", c.Log.Format))
	}
	if c.Userkit.Enabled {
		if strings.TrimSpace(c.Userkit.DatabaseURL) == "" {
			errs = append(errs, errors.New("USERKIT_DATABASE_URL is required when USERKIT_ENABLED=true"))
		}
		if len(c.Userkit.JWTSecret) < 16 {
			errs = append(errs, errors.New("USERKIT_JWT_SECRET must be at least 16 characters when USERKIT_ENABLED=true"))
		}
		if c.Userkit.TokenTTL <= 0 {
			errs = append(errs, errors.New("USERKIT_TOKEN_TTL must be positive"))
		}
		if c.Userkit.DBMaxOpenConns <= 0 {
			errs = append(errs, errors.New("USERKIT_DB_MAX_OPEN_CONNS must be positive"))
		}
		if c.Userkit.DBMaxIdleConns <= 0 {
			errs = append(errs, errors.New("USERKIT_DB_MAX_IDLE_CONNS must be positive"))
		}
		if c.Userkit.DBConnMaxLifetime <= 0 {
			errs = append(errs, errors.New("USERKIT_DB_CONN_MAX_LIFETIME must be positive"))
		}
		if c.Userkit.DBConnMaxIdleTime <= 0 {
			errs = append(errs, errors.New("USERKIT_DB_CONN_MAX_IDLE_TIME must be positive"))
		}
	}

	return errors.Join(errs...)
}

func envString(lookup LookupFunc, key, fallback string) string {
	value, ok := lookup(key)
	if !ok {
		return fallback
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func envDuration(lookup LookupFunc, key string, fallback time.Duration) time.Duration {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return -1
	}
	return parsed
}

func envInt(lookup LookupFunc, key string, fallback int) int {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return -1
	}
	return parsed
}

func envBool(lookup LookupFunc, key string, fallback bool) bool {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func envCSV(lookup LookupFunc, key string) []string {
	value, ok := lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func validLogLevel(level string) bool {
	switch level {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}
