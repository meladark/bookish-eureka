package main

import (
	"github.com/spf13/viper"
)

type Config struct {
	Logger struct {
		Level   string `mapstructure:"level"`
		Logfile string `mapstructure:"logfile"`
	} `mapstructure:"logger"`
	HTTPServer struct {
		Host string `mapstructure:"host"`
		Port string `mapstructure:"port"`
	} `mapstructure:"http_server"`
	Storage struct {
		Type string `mapstructure:"type"` // memory | sql
		SQL  struct {
			DSN string `mapstructure:"dsn"`
		} `mapstructure:"sql"`
	} `mapstructure:"storage"`
}

func NewConfig(configFile string) (*Config, error) {
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
