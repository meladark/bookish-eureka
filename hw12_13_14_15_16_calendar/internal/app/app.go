package app

import (
	"context"
	"fmt"
	"time"

	logger "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	logger  logger.Logger
	storage storage.Storage
}

func New(logger logger.Logger, storage storage.Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, e storage.Event) error {
	e.StartTime = e.StartTime.UTC()
	e.EndTime = e.EndTime.UTC()
	a.logger.Info(fmt.Sprintf("CreateEvent: %s (start: %s)", e.ID, e.StartTime))
	return a.storage.CreateEvent(ctx, e)
}

func (a *App) UpdateEvent(ctx context.Context, e storage.Event) error {
	e.StartTime = e.StartTime.UTC()
	e.EndTime = e.EndTime.UTC()
	a.logger.Info(fmt.Sprintf("UpdateEvent: %s (start: %s)", e.ID, e.StartTime))
	return a.storage.UpdateEvent(ctx, e)
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	a.logger.Info("DeleteEvent: " + id)
	return a.storage.DeleteEvent(ctx, id)
}

func (a *App) GetEvent(ctx context.Context, id string) (*storage.Event, error) {
	a.logger.Info("GetEvent: " + id)
	event, err := a.storage.GetEvent(ctx, id)
	if err != nil {
		return nil, err
	}
	if event != nil {
		event.StartTime = event.StartTime.UTC()
		event.EndTime = event.EndTime.UTC()
	}
	return event, nil
}

func (a *App) ListEventsByDay(ctx context.Context, day time.Time) ([]storage.Event, error) {
	all, err := a.storage.ListEvents(ctx)
	if err != nil {
		return nil, err
	}
	start := day.UTC().Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	a.logger.Debug(fmt.Sprintf("ListEventsByDay range: %s to %s (UTC)", start, end))
	return filterByTimeRange(all, start, end), nil
}

func (a *App) ListEventsByWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error) {
	all, err := a.storage.ListEvents(ctx)
	if err != nil {
		return nil, err
	}
	start := weekStart.UTC().Truncate(24 * time.Hour)
	for start.Weekday() != time.Monday {
		start = start.AddDate(0, 0, -1)
	}
	end := start.AddDate(0, 0, 7)
	a.logger.Debug(fmt.Sprintf("ListEventsByWeek range: %s to %s (UTC)", start, end))
	return filterByTimeRange(all, start, end), nil
}

func (a *App) ListEventsByMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error) {
	all, err := a.storage.ListEvents(ctx)
	if err != nil {
		return nil, err
	}
	start := time.Date(
		monthStart.UTC().Year(),
		monthStart.UTC().Month(),
		1, 0, 0, 0, 0, time.UTC,
	)
	end := start.AddDate(0, 1, 0)
	a.logger.Debug(fmt.Sprintf("ListEventsByMonth range: %s to %s (UTC)", start, end))
	return filterByTimeRange(all, start, end), nil
}

func filterByTimeRange(events []storage.Event, start, end time.Time) []storage.Event {
	var result []storage.Event
	for _, e := range events {
		st := e.StartTime.UTC()
		if !st.Before(start) && st.Before(end) {
			filteredEvent := e
			filteredEvent.StartTime = st
			filteredEvent.EndTime = e.EndTime.UTC()
			result = append(result, filteredEvent)
		}
	}
	return result
}
