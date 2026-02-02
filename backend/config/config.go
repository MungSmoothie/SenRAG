package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	LLM      LLMConfig      `yaml:"llm"`
	Qdrant   QdrantConfig   `yaml:"qdrant"`
	Upload   UploadConfig   `yaml:"upload"`
	Embedder EmbedderConfig `yaml:"embedder"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LLMConfig struct {
	Provider   string `yaml:"provider"` // openai, ollama, local
	BaseURL    string `yaml:"base_url"`
	APIKey     string `yaml:"api_key"`
	Model      string `yaml:"model"`
	MaxTokens  int    `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

type QdrantConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	APIKey   string `yaml:"api_key"`
	UseAuth  bool   `yaml:"use_auth"`
	CollectionName string `yaml:"collection_name"`
}

type UploadConfig struct {
	MaxFileSize int64  `yaml:"max_file_size"`
	AllowedExts []string `yaml:"allowed_extensions"`
	SavePath    string `yaml:"save_path"`
}

type EmbedderConfig struct {
	Model string `yaml:"model"`
	BatchSize int `yaml:"batch_size"`
}

var globalConfig *Config

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 加载环境变量
	cfg.applyEnvVars()

	// 默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Qdrant.Port == 0 {
		cfg.Qdrant.Port = 6333
	}
	if cfg.Qdrant.CollectionName == "" {
		cfg.Qdrant.CollectionName = "senrag_collection"
	}
	if cfg.Upload.MaxFileSize == 0 {
		cfg.Upload.MaxFileSize = 50 << 20 // 50MB
	}
	if cfg.Upload.SavePath == "" {
		cfg.Upload.SavePath = "./uploads"
	}
	if cfg.Embedder.BatchSize == 0 {
		cfg.Embedder.BatchSize = 32
	}

	globalConfig = &cfg
	return &cfg, nil
}

func GetConfig() *Config {
	return globalConfig
}

func (c *Config) applyEnvVars() {
	if v := os.Getenv("LLM_BASE_URL"); v != "" {
		c.LLM.BaseURL = v
	}
	if v := os.Getenv("LLM_API_KEY"); v != "" {
		c.LLM.APIKey = v
	}
	if v := os.Getenv("LLM_MODEL"); v != "" {
		c.LLM.Model = v
	}
	if v := os.Getenv("QDRANT_HOST"); v != "" {
		c.Qdrant.Host = v
	}
	if v := os.Getenv("QDRANT_API_KEY"); v != "" {
		c.Qdrant.APIKey = v
	}
}

func GetConfigPath() string {
	paths := []string{
		"config.yaml",
		".config.yaml",
		".env.yaml",
		"/etc/senrag/config.yaml",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			return abs
		}
	}
	return "config.yaml"
}
