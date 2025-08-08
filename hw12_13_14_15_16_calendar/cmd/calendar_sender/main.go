package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	logger "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/messaging"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/pkg/models"
	"github.com/spf13/viper"
)

type Config struct {
	Logger struct {
		Level   string `mapstructure:"level"`
		Logfile string `mapstructure:"logfile"`
	} `mapstructure:"logger"`
	RabbitMQ struct {
		URL       string `mapstructure:"url"`
		QueueName string `mapstructure:"queue_name"`
	} `mapstructure:"rabbitmq"`
}

func main() {
	configPath := getConfigPath()
	viper.SetConfigFile(configPath)
	viper.SetEnvPrefix("SENDER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
	logg := logger.New(cfg.Logger.Level, cfg.Logger.Logfile)
	rmq, err := messaging.New(cfg.RabbitMQ.URL, cfg.RabbitMQ.QueueName)
	if err != nil {
		logg.Error("failed to connect to RabbitMQ: " + err.Error())
	}
	defer rmq.Close()

	msgs, err := rmq.Consume()
	if err != nil {
		logg.Error("failed to consume: " + err.Error())
	}

	logg.Info("waiting for notifications...")

	for msg := range msgs {
		var n models.Notification
		if err := json.Unmarshal(msg.Body, &n); err != nil {
			log.Printf("bad message: %v", err)
			continue
		}
		logg.Info("Notify: " + n.EventID + " Title: " + n.Title)
		if err := msg.Ack(false); err != nil {
			logg.Error("failed to ack message: " + err.Error())
		}
	}
}

func getConfigPath() string {
	for i, arg := range os.Args {
		if arg == "--config" && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
}
