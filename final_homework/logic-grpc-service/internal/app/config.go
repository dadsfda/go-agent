package app

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "config/config.yaml"

type AppConfig struct {
	MySQL struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`
	COS COSConfig `yaml:"cos"`
	AI  AIConfig  `yaml:"ai"`
}

func LoadConfig(path string) (AppConfig, error) {
	var cfg AppConfig
	if path == "" {
		path = findConfigPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return AppConfig{}, err
	}
	return cfg, nil
}

func findConfigPath() string {
	candidates := []string{
		DefaultConfigPath,
		filepath.Join("..", DefaultConfigPath),
		filepath.Join("..", "..", DefaultConfigPath),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return DefaultConfigPath
}

func ConfigFromFileAndEnv(path string) (AppConfig, error) {
	cfg, err := LoadConfig(path)
	if err != nil {
		return AppConfig{}, err
	}
	if value := os.Getenv("MYSQL_DSN"); value != "" {
		cfg.MySQL.DSN = value
	}
	if value := os.Getenv("COS_BUCKET"); value != "" {
		cfg.COS.Bucket = value
	}
	if value := os.Getenv("COS_REGION"); value != "" {
		cfg.COS.Region = value
	}
	if value := os.Getenv("COS_ENDPOINT"); value != "" {
		cfg.COS.Endpoint = value
	}
	if value := os.Getenv("TENCENT_SECRET_ID"); value != "" {
		cfg.COS.SecretID = value
	}
	if value := os.Getenv("TENCENT_SECRET_KEY"); value != "" {
		cfg.COS.SecretKey = value
	}
	if value := os.Getenv("DEEPSEEK_API_KEY"); value != "" {
		cfg.AI.APIKey = value
	}
	if value := os.Getenv("DEEPSEEK_MODEL"); value != "" {
		cfg.AI.Model = value
	}
	if value := os.Getenv("DEEPSEEK_BASE_URL"); value != "" {
		cfg.AI.BaseURL = value
	}
	return cfg, nil
}
