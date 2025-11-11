package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	HTTPHost string
	HTTPPort string
	staticFilesPath string
}

func LoadConfig() (*Config, error) {
	httpHost := os.Getenv("httpHost")
	if strings.TrimSpace(httpHost) == "" {
		return nil, fmt.Errorf("httpHost must be set, check .env file")
	}

	httpPort := os.Getenv("httpPort")
	if strings.TrimSpace(httpPort) == "" {
		return nil, fmt.Errorf("httpPort must be set, check .env file")
	}

	staticFilesPath := os.Getenv("staticFilesPath")
	fmt.Println(staticFilesPath)
	if strings.TrimSpace(staticFilesPath) == "" {
		return nil, fmt.Errorf("staticFilesPath must be set, check .env file")
	}

	return &Config{
		HTTPHost: httpHost,
		HTTPPort: httpPort,
		staticFilesPath: staticFilesPath,
	}, nil
}


func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.HTTPHost, c.HTTPPort)
}

func (c *Config) GetStaticFilesPath() string {
	return c.staticFilesPath
}

