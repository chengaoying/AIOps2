package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port        string            `yaml:"port"`
	StarRocks   StarRocksConfig   `yaml:"starrocks"`
	Plugins     map[string]Plugin `yaml:"plugins"`
	WAL         WALConfig         `yaml:"wal"`
	BatchWriter BatchConfig       `yaml:"batch_writer"`
}

type StarRocksConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type Plugin struct {
	Enabled  bool          `yaml:"enabled"`
	APIURL   string        `yaml:"api_url"`
	Interval time.Duration `yaml:"interval"`
	Extra    map[string]any `yaml:"extra,omitempty"`
}

type WALConfig struct {
	Dir         string `yaml:"dir"`
	MaxFileSize int64 `yaml:"max_file_size"`
}

type BatchConfig struct {
	BatchSize     int           `yaml:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Port == "" {
		cfg.Port = "8081"
	}
	if cfg.StarRocks.Port == 0 {
		cfg.StarRocks.Port = 9030
	}
	if cfg.WAL.Dir == "" {
		cfg.WAL.Dir = "/var/log/aiops/wal"
	}
	if cfg.WAL.MaxFileSize == 0 {
		cfg.WAL.MaxFileSize = 1 << 30 // 1GB
	}
	if cfg.BatchWriter.BatchSize == 0 {
		cfg.BatchWriter.BatchSize = 1000
	}
	if cfg.BatchWriter.FlushInterval == 0 {
		cfg.BatchWriter.FlushInterval = 5 * time.Second
	}
}

func (c *StarRocksConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4",
		c.User, c.Password, c.Host, c.Port, c.Database)
}
