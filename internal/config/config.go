package config

import (
	"errors"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Port string `yaml:"port"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name   string `yaml:"name"`
}

type Config struct {
	HTTPServer HTTPServer `yaml:"http_server"`
	Database Database `yaml:"database"`
}

func MustLoad() (Config, error) {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		return Config{}, errors.New("no config path provided")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, errors.New("no config file provided")
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
