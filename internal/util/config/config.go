package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration
type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Kafka          KafkaConfig          `mapstructure:"kafka"`
	Redis          RedisConfig          `mapstructure:"redis"`
	Postgres       PostgresConfig       `mapstructure:"postgres"`
	Processing     ProcessingConfig     `mapstructure:"processing"`
	Recommendation RecommendationConfig `mapstructure:"recommendation"`
	EventWeights   EventWeightsConfig   `mapstructure:"event_weights"`
	Observability  ObservabilityConfig  `mapstructure:"observability"`
}

type ServerConfig struct {
	Ingest APIServerConfig `mapstructure:"ingest"`
	API    APIServerConfig `mapstructure:"api"`
}

type APIServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type KafkaConfig struct {
	Brokers       []string    `mapstructure:"brokers"`
	Topics        TopicConfig `mapstructure:"topics"`
	ConsumerGroup string      `mapstructure:"consumer_group"`
}

type TopicConfig struct {
	Events string `mapstructure:"events"`
}

type RedisConfig struct {
	Addr       string `mapstructure:"addr"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

type PostgresConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Database, p.SSLMode)
}

type ProcessingConfig struct {
	BatchSize        int           `mapstructure:"batch_size"`
	FlushInterval    time.Duration `mapstructure:"flush_interval"`
	RecentItemsLimit int           `mapstructure:"recent_items_limit"`
	CoviewWindow     int           `mapstructure:"coview_window"`
}

type RecommendationConfig struct {
	DefaultCount    int           `mapstructure:"default_count"`
	MaxCount        int           `mapstructure:"max_count"`
	CacheTTL        time.Duration `mapstructure:"cache_ttl"`
	PopularityDecay float64       `mapstructure:"popularity_decay"`
	Weights         WeightsConfig `mapstructure:"weights"`
}

type WeightsConfig struct {
	Coview     float64 `mapstructure:"coview"`
	Embedding  float64 `mapstructure:"embedding"`
	Popularity float64 `mapstructure:"popularity"`
	Recency    float64 `mapstructure:"recency"`
}

type EventWeightsConfig struct {
	View     float64 `mapstructure:"VIEW"`
	Click    float64 `mapstructure:"CLICK"`
	Cart     float64 `mapstructure:"CART"`
	Purchase float64 `mapstructure:"PURCHASE"`
}

type ObservabilityConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
	Tracing TracingConfig `mapstructure:"tracing"`
	Logging LoggingConfig `mapstructure:"logging"`
}

type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

type TracingConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load loads configuration from file and environment
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath("../config")
		v.AddConfigPath("../../config")
	}

	// Read from environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("RECO")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
