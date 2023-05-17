// config/config.go
package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type ClickhouseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type Config struct {
	NodeURL    string           `yaml:"node_url"`
	HTTPURL    string           `yaml:"http_url"`
	Clickhouse ClickhouseConfig `yaml:"clickhouse"`
}

func LoadConfig() Config {
	data, err := ioutil.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	return cfg
}
