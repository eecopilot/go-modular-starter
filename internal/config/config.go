package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultDevelopmentJWTSecret = "change-me-to-a-long-random-secret"

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
	p := &envParser{lookup: lookup}

	cfg := Config{
		App: AppConfig{
			Name:      p.str("APP_NAME", "go-modular-starter"),
			Env:       p.str("APP_ENV", "development"),
			Version:   p.str("APP_VERSION", "dev"),
			Commit:    p.str("APP_COMMIT", "unknown"),
			BuildTime: p.str("APP_BUILD_TIME", "unknown"),
		},
		HTTP: HTTPConfig{
			Addr:               p.str("HTTP_ADDR", ":8080"),
			ReadHeaderTimeout:  p.duration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
			ReadTimeout:        p.duration("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:       p.duration("HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:        p.duration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout:    p.duration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
			MaxHeaderBytes:     p.integer("HTTP_MAX_HEADER_BYTES", 1<<20),
			BodyLimitBytes:     int64(p.integer("HTTP_BODY_LIMIT_BYTES", 1<<20)),
			CORSAllowedOrigins: p.csv("HTTP_CORS_ALLOWED_ORIGINS"),
		},
		Log: LogConfig{
			Level:  strings.ToLower(p.str("LOG_LEVEL", "info")),
			Format: strings.ToLower(p.str("LOG_FORMAT", "text")),
		},
		Userkit: UserkitConfig{
			Enabled:           p.boolean("USERKIT_ENABLED", false),
			DatabaseURL:       p.str("USERKIT_DATABASE_URL", ""),
			DBMaxOpenConns:    p.integer("USERKIT_DB_MAX_OPEN_CONNS", 10),
			DBMaxIdleConns:    p.integer("USERKIT_DB_MAX_IDLE_CONNS", 5),
			DBConnMaxLifetime: p.duration("USERKIT_DB_CONN_MAX_LIFETIME", 30*time.Minute),
			DBConnMaxIdleTime: p.duration("USERKIT_DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
			JWTSecret:         p.str("USERKIT_JWT_SECRET", ""),
			JWTIssuer:         p.str("USERKIT_JWT_ISSUER", "go-modular-starter"),
			JWTAudience:       p.str("USERKIT_JWT_AUDIENCE", "go-modular-starter-users"),
			TokenTTL:          p.duration("USERKIT_TOKEN_TTL", 24*time.Hour),
			BCryptCost:        p.integer("USERKIT_BCRYPT_COST", 0),
		},
	}

	// 解析错误必须直接失败：静默回退默认值会掩盖 USERKIT_ENABLED=ture 这类手误。
	if err := errors.Join(p.errs...); err != nil {
		return Config{}, err
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
		if isProductionEnv(c.App.Env) && isDefaultJWTSecret(c.Userkit.JWTSecret) {
			errs = append(errs, errors.New("USERKIT_JWT_SECRET must be changed in production"))
		}
		if c.Userkit.TokenTTL <= 0 {
			errs = append(errs, errors.New("USERKIT_TOKEN_TTL must be positive"))
		}
		// 0 表示交给 userkit 用 bcrypt.DefaultCost；显式设置时限定 bcrypt 合法区间。
		if c.Userkit.BCryptCost != 0 && (c.Userkit.BCryptCost < 4 || c.Userkit.BCryptCost > 31) {
			errs = append(errs, errors.New("USERKIT_BCRYPT_COST must be 0 (default) or between 4 and 31"))
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

func isProductionEnv(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "prod", "production":
		return true
	default:
		return false
	}
}

func isDefaultJWTSecret(secret string) bool {
	return strings.TrimSpace(secret) == defaultDevelopmentJWTSecret
}

type envParser struct {
	lookup LookupFunc
	errs   []error
}

func (p *envParser) str(key, fallback string) string {
	value, ok := p.lookup(key)
	if !ok {
		return fallback
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func (p *envParser) duration(key string, fallback time.Duration) time.Duration {
	value, ok := p.lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		p.errs = append(p.errs, fmt.Errorf("%s: invalid duration %q (example: 15s, 1m30s)", key, value))
		return fallback
	}
	return parsed
}

func (p *envParser) integer(key string, fallback int) int {
	value, ok := p.lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		p.errs = append(p.errs, fmt.Errorf("%s: invalid integer %q", key, value))
		return fallback
	}
	return parsed
}

func (p *envParser) boolean(key string, fallback bool) bool {
	value, ok := p.lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		p.errs = append(p.errs, fmt.Errorf("%s: invalid boolean %q (use true or false)", key, value))
		return fallback
	}
	return parsed
}

func (p *envParser) csv(key string) []string {
	value, ok := p.lookup(key)
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
