package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for TelumDB
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	Logging LoggingConfig `yaml:"logging"`
	Metrics MetricsConfig `yaml:"metrics"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	HTTPPort        int           `yaml:"http_port"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	MaxConnections  int           `yaml:"max_connections"`
	EnableTLS       bool          `yaml:"enable_tls"`
	CertFile        string        `yaml:"cert_file"`
	KeyFile         string        `yaml:"key_file"`
}

// StorageConfig contains storage-related configuration
type StorageConfig struct {
	DataDir            string        `yaml:"data_dir"`
	Engine             string        `yaml:"engine"`
	MaxFileSize        int64         `yaml:"max_file_size"`
	Compression        string        `yaml:"compression"`
	CacheSize          int64         `yaml:"cache_size"`
	SyncMode           string        `yaml:"sync_mode"`
	WALEnabled         bool          `yaml:"wal_enabled"`
	CheckpointInterval time.Duration `yaml:"checkpoint_interval"`
	TensorConfig       TensorConfig  `yaml:"tensor"`
}

// TensorConfig contains tensor-specific configuration
type TensorConfig struct {
	ChunkSize      []int  `yaml:"chunk_size"`
	DefaultDType   string `yaml:"default_dtype"`
	Compression    string `yaml:"compression"`
	MemoryLimit    int64  `yaml:"memory_limit"`
	Parallelism    int    `yaml:"parallelism"`
	GPUEnabled     bool   `yaml:"gpu_enabled"`
	GPUMemoryLimit int64  `yaml:"gpu_memory_limit"`
}

// LoggingConfig contains logging-related configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// MetricsConfig contains metrics-related configuration
type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Port      int    `yaml:"port"`
	Path      string `yaml:"path"`
	Namespace string `yaml:"namespace"`
}

// Load loads configuration from file or environment variables
func Load(configFile string) (*Config, error) {
	cfg := &Config{}

	// Set defaults
	setDefaults(cfg)

	// Load from file if provided
	if configFile != "" {
		if err := loadFromFile(cfg, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	} else {
		// Try default locations
		defaultFiles := []string{
			"config.yaml",
			"config.yml",
			"/etc/telumdb/config.yaml",
			os.Getenv("HOME") + "/.telumdb/config.yaml",
		}

		for _, file := range defaultFiles {
			if _, err := os.Stat(file); err == nil {
				if err := loadFromFile(cfg, file); err != nil {
					return nil, fmt.Errorf("failed to load config from file %s: %w", file, err)
				}
				break
			}
		}
	}

	// Override with environment variables
	loadFromEnv(cfg)

	// Validate configuration
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// setDefaults sets default values for configuration
func setDefaults(cfg *Config) {
	cfg.Server = ServerConfig{
		Host:            "0.0.0.0",
		Port:            5432,
		HTTPPort:        8080,
		ShutdownTimeout: 30 * time.Second,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		MaxConnections:  1000,
		EnableTLS:       false,
	}

	cfg.Storage = StorageConfig{
		DataDir:            "./data",
		Engine:             "hybrid",
		MaxFileSize:        1 << 30, // 1GB
		Compression:        "lz4",
		CacheSize:          1 << 30, // 1GB
		SyncMode:           "normal",
		WALEnabled:         true,
		CheckpointInterval: 5 * time.Minute,
		TensorConfig: TensorConfig{
			ChunkSize:      []int{64, 64, 64},
			DefaultDType:   "float32",
			Compression:    "zstd",
			MemoryLimit:    4 << 30, // 4GB
			Parallelism:    4,
			GPUEnabled:     false,
			GPUMemoryLimit: 2 << 30, // 2GB
		},
	}

	cfg.Logging = LoggingConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100, // MB
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	}

	cfg.Metrics = MetricsConfig{
		Enabled:   true,
		Port:      9000,
		Path:      "/metrics",
		Namespace: "telumdb",
	}
}

// loadFromFile loads configuration from YAML file
func loadFromFile(cfg *Config, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(cfg *Config) {
	if host := os.Getenv("TELUMDB_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("TELUMDB_PORT"); port != "" {
		if p, err := parseInt(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if dataDir := os.Getenv("TELUMDB_DATA_DIR"); dataDir != "" {
		cfg.Storage.DataDir = dataDir
	}
	if logLevel := os.Getenv("TELUMDB_LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}
}

// validate validates the configuration
func validate(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}
	if cfg.Server.HTTPPort <= 0 || cfg.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", cfg.Server.HTTPPort)
	}
	if cfg.Storage.DataDir == "" {
		return fmt.Errorf("storage data directory cannot be empty")
	}
	if cfg.Storage.CacheSize <= 0 {
		return fmt.Errorf("storage cache size must be positive")
	}
	return nil
}

// parseInt parses string to int with error handling
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
