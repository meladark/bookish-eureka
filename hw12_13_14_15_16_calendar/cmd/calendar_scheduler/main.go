package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	logger "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/messaging"
	sql "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/sql"
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
	Database struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"database"`
	Scheduler struct {
		Interval string `mapstructure:"interval"`
	} `mapstructure:"scheduler"`
}

func main() {
	configPath := getConfigPath()
	viper.SetConfigFile(configPath)
	viper.SetEnvPrefix("SCHEDULER")
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
	defer logg.Close()
	logg.Info(cfg.Database.DSN)
	db, err := sql.New(cfg.Database.DSN)
	if err != nil {
		logg.Error("failed to init storage: " + err.Error())
	}

	rmq, err := messaging.New(cfg.RabbitMQ.URL, cfg.RabbitMQ.QueueName)
	if err != nil {
		logg.Error("failed to connect to RabbitMQ: " + err.Error())
	}
	defer rmq.Close()

	interval, _ := time.ParseDuration(cfg.Scheduler.Interval)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	for {
		now := time.Now()
		events, err := db.EventsToNotify(ctx, now)
		if err != nil {
			logg.Info("error selecting events: " + err.Error())
			continue
		}

		for _, ev := range events {
			notification := models.Notification{
				EventID:     ev.ID,
				Title:       ev.Title,
				ScheduledAt: ev.NotifyAt,
			}
			body, _ := json.Marshal(notification)
			if err := rmq.Publish(body); err != nil {
				logg.Info("error publishing: " + err.Error())
			} else {
				logg.Info("notification sent: " + ev.ID)
			}
		}
		err = db.DeleteOldEvents(ctx, now.AddDate(-1, 0, 0))
		if err != nil {
			logg.Info("error cleaning old events: " + err.Error())
		}

		time.Sleep(interval)
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
