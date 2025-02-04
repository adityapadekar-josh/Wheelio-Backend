package config

import (
	"errors"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Port string `yaml:"port" env:"HTTP_PORT" required:"true"`
}

type Database struct {
	Host     string `yaml:"host" required:"true"`
	Port     int    `yaml:"port" required:"true"`
	User     string `yaml:"user" required:"true"`
	Password string `yaml:"password" required:"true"`
	Name     string `yaml:"name" required:"true"`
}

type EmailService struct {
	ApiKey    string `yaml:"api_key" required:"true"`
	FromName  string `yaml:"from_name" required:"true"`
	FromEmail string `yaml:"from_email" required:"true"`
}

type Config struct {
	HTTPServer   HTTPServer   `yaml:"http_server"`
	Database     Database     `yaml:"database"`
	EmailService EmailService `yaml:"email_service"`
	JWTSecret    string       `yaml:"jwt_secret"`
}

var cfg Config

func MustLoad() (Config, error) {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		return Config{}, errors.New("no config path provided")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, errors.New("no config file provided")
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func GetConfig() Config {
	return cfg
}
