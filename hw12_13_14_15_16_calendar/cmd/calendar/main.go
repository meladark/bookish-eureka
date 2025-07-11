package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	//nolint:depguard
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	//nolint:depguard
	logger "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	//nolint:depguard
	internalhttp "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/http"
	//nolint:depguard
	memorystorage "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/memory"
	//nolint:depguard
	sql "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := NewConfig(configFile)
	if err != nil {
		panic("cannot load config: " + err.Error())
	}
	logg := logger.New(cfg.Logger.Level, cfg.Logger.Logfile)

	var storage app.Storage

	switch cfg.Storage.Type {
	case "memory":
		storage = memorystorage.New()
	case "sql":
		db, err := sql.New(cfg.Storage.SQL.DSN)
		if err != nil {
			logg.Error("cannot init sql storage: " + err.Error())
			os.Exit(1)
		}
		storage = db
	default:
		logg.Error("unknown storage type: " + cfg.Storage.Type)
		os.Exit(1)
	}
	_ = app.New(logg, storage)

	server := internalhttp.NewServer(logg, cfg.HTTPServer.Host, cfg.HTTPServer.Port)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := server.Stop(ctxTimeout); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		//nolint:gocritic // reason: выше явно есть cancel
		os.Exit(1)
	}
}
