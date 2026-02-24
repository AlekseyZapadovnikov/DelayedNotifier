package config

import "fmt"

type HTTPConfig struct {
	Host            string 	
	Port            string `env:"httpPort" env-required:"true" validate:"required,port"`
	StaticFilesPath string `env:"staticFilesPath" env-required:"true" validate:"required"`
}

func (c *HTTPConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func (c *HTTPConfig) GetServerAddress() string {
	return c.Address()
}

func (c *HTTPConfig) GetStaticFilesPath() string {
	return c.StaticFilesPath
}
