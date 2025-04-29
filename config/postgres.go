package config

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// PostgresConfig defines the configuration for connecting to a PostgreSQL database.
type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	TimeZone string `mapstructure:"timezone"`

	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

func (cfg *PostgresConfig) DSN(env string) string {
	if env == "prod" {
		dbHost := getParameterStoreValue("DJANGO_DEFAULT_DB_HOST", true)
		dbUser := getParameterStoreValue("DJANGO_DEFAULT_DB_USER", true)
		dbPassword := getParameterStoreValue("DJANGO_DEFAULT_DB_PASSWORD", true)

		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dbHost, cfg.Port, dbUser, dbPassword, cfg.DBName, cfg.SSLMode,
		)

		if cfg.TimeZone != "" {
			dsn += fmt.Sprintf(" TimeZone=%s", cfg.TimeZone)
		}

		return dsn
	} else {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		)

		if cfg.TimeZone != "" {
			dsn += fmt.Sprintf(" TimeZone=%s", cfg.TimeZone)
		}

		return dsn
	}
}

func getParameterStoreValue(parameterName string, decrypt bool) string {
	baseCtx := context.Background()
	ctxWithTimeout, cancel := context.WithTimeout(baseCtx, 5*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctxWithTimeout)
	if err != nil {
		return ""
	}

	client := ssm.NewFromConfig(cfg)

	input := &ssm.GetParameterInput{
		Name:           &parameterName,
		WithDecryption: &decrypt,
	}

	result, err := client.GetParameter(ctxWithTimeout, input)
	if err != nil {
		return ""
	}

	if result.Parameter.Value == nil {
		return ""
	}

	return *result.Parameter.Value
}
