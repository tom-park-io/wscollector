package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Bybit    BybitConfig    `mapstructure:"bybit"`
	Log      LogConfig      `mapstructure:"log"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type BybitConfig struct {
	REST RESTConfig `mapstructure:"rest"`
	WS   WSConfig   `mapstructure:"ws"`
}

type RESTConfig struct {
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
}
type WSConfig struct {
	URL      string        `mapstructure:"url"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Interval string        `mapstructure:"interval"`
}

// Options defines the logger configuration options.
type LogConfig struct {
	Level       string `mapstructure:"level"`       // log level: "debug", "info", "warn", "error"
	Format      string `mapstructure:"format"`      // log format: "json" or "console"
	OutputFile  string `mapstructure:"output_file"` // file path to store logs (optional)
	Environment string `mapstructure:"environment"` // environment: "dev" or "prod"
}

// Load loads application configuration using Viper.
// It reads from config.yaml and overrides with environment variables.
func Load() *Config {
	v := viper.New()

	v.SetConfigName("config") // config.yaml
	v.SetConfigType("yaml")

	// TODO: env path
	ex, _ := os.Executable()
	if strings.Contains(ex, "go-build") {
		pwd, _ := os.Getwd()
		v.AddConfigPath(filepath.Join(pwd, "../../config"))
	} else {
		v.AddConfigPath(filepath.Join(filepath.Dir(ex), "../config"))
	}

	// Support environment variables with dot notation (e.g., BYBIT_WS_URL)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	return &cfg
}
