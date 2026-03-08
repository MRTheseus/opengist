package config

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

var C *Config

type Config struct {
	SecretKey    string
	LogLevel     string
	ExternalUrl  string
	DataDir      string
	DBPath       string
	HttpHost     string
	HttpPort     string
	CustomName   string
}

func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dataDir := getEnv("OG_DATA_DIR", filepath.Join(homeDir, ".closegist"))
	
	C = &Config{
		SecretKey:   getEnv("OG_SECRET_KEY", "change-this-secret-key"),
		LogLevel:    getEnv("OG_LOG_LEVEL", "warn"),
		ExternalUrl: getEnv("OG_EXTERNAL_URL", ""),
		DataDir:     dataDir,
		DBPath:      getEnv("OG_DB_PATH", filepath.Join(dataDir, "closegist.db")),
		HttpHost:    getEnv("OG_HTTP_HOST", "0.0.0.0"),
		HttpPort:    getEnv("OG_HTTP_PORT", "6157"),
		CustomName: getEnv("OG_CUSTOM_NAME", "CloseGist"),
	}

	if err = os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	zerolog.SetGlobalLevel(getLogLevel(C.LogLevel))

	return nil
}

func GetDataDir() string {
	return C.DataDir
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.WarnLevel
	}
}
