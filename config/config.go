package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server    ServerConfig
	MongoDB   MongoDBConfig
	Logging   LoggingConfig
	Security  SecurityConfig
	Dashboard DashboardConfig
	Kafka     KafkaConfig
	JWT       JWTConfig
	Rendering RenderingConfig
	Storage   StorageConfig
}

type KafkaConfig struct {
	Enabled        bool
	Brokers        []string
	Topics         []string
	GroupID        string
	ClientID       string
	AutoCommit     bool
	AutoOffset     string
	SessionTimeout int
	HeartbeatInt   int
}

type ServerConfig struct {
	Port string `json:"port"`
}

type MongoDBConfig struct {
	URI      string `json:"mongodb_uri"`
	Database string `json:"mongodb_database"`
}

type LoggingConfig struct {
	Level  string `json:"log_level"`
	Format string `json:"log_format"`
}

type SecurityConfig struct {
	AdminAPIKey     string `json:"admin_api_key"`
	RateLimitPerMin int    `json:"rate_limit_per_minute"`
	APIKeyEnabled   bool   `json:"api_key_enabled"`
}

type DashboardConfig struct {
	Enabled  bool   `json:"dashboard_enabled"`
	Username string `json:"dashboard_username"`
	Password string `json:"dashboard_password"`
}

type JWTConfig struct {
	Secret string `json:"jwt_secret"`
}

type RenderingConfig struct {
	ChromePath    string `json:"chrome_path"`
	DefaultWidth  int    `json:"default_width"`
	DefaultHeight int    `json:"default_height"`
	DefaultFormat string `json:"default_format"`
	DefaultQuality int   `json:"default_quality"`
}

type StorageConfig struct {
	OutputDir string `json:"output_dir"`
	BaseURL   string `json:"base_url"`
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "9104"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://app_user:app_pass@localhost:27022/imposizcy?authSource=test"),
			Database: getEnv("MONGODB_DATABASE", "imposizcy"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Security: SecurityConfig{
			AdminAPIKey:     getEnv("ADMIN_API_KEY", "imposizcy-dev-key-2024"),
			RateLimitPerMin: getEnvInt("RATE_LIMIT_PER_MINUTE", 60),
			APIKeyEnabled:   getEnvBool("API_KEY_ENABLED", false),
		},
		Dashboard: DashboardConfig{
			Enabled:  getEnvBool("DASHBOARD_ENABLED", true),
			Username: getEnv("DASHBOARD_USERNAME", "admin"),
			Password: getEnv("DASHBOARD_PASSWORD", "admin123"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "imposizcy-jwt-secret-dev"),
		},
		Rendering: RenderingConfig{
			ChromePath:     getEnv("CHROME_PATH", ""),
			DefaultWidth:   getEnvInt("DEFAULT_RENDER_WIDTH", 1080),
			DefaultHeight:  getEnvInt("DEFAULT_RENDER_HEIGHT", 1080),
			DefaultFormat:  getEnv("DEFAULT_RENDER_FORMAT", "png"),
			DefaultQuality: getEnvInt("DEFAULT_RENDER_QUALITY", 90),
		},
		Storage: StorageConfig{
			OutputDir: getEnv("OUTPUT_DIR", "./data/outputs"),
			BaseURL:   getEnv("OUTPUT_BASE_URL", ""),
		},
		Kafka: KafkaConfig{
			Enabled:        getEnvBool("KAFKA_ENABLED", false),
			Brokers:        getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			Topics:         getEnvSlice("KAFKA_TOPICS", []string{"imposizcy-render"}),
			GroupID:        getEnv("KAFKA_GROUP_ID", "imposizcy-group"),
			ClientID:       getEnv("KAFKA_CLIENT_ID", "imposizcy"),
			AutoCommit:     getEnvBool("KAFKA_AUTO_COMMIT", false),
			AutoOffset:     getEnv("KAFKA_AUTO_OFFSET", "earliest"),
			SessionTimeout: getEnvInt("KAFKA_SESSION_TIMEOUT", 10000),
			HeartbeatInt:   getEnvInt("KAFKA_HEARTBEAT_INTERVAL", 3000),
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if val == "true" || val == "1" || val == "yes" {
			return true
		}
		if val == "false" || val == "0" || val == "no" {
			return false
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if val := os.Getenv(key); val != "" {
		var result []string
		for _, s := range splitByComma(val) {
			if trimmed := trimSpace(s); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return fallback
}

func splitByComma(s string) []string {
	return strings.Split(s, ",")
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
