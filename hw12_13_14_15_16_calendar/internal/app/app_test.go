package app_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	event "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestApp(t *testing.T) *app.App {
	t.Helper()
	logFile := "./test.log"
	log := logger.New("debug", logFile)
	t.Cleanup(func() { os.Remove(logFile) })
	return app.New(log, memorystorage.New())
}

func createTestEvent(id string, start time.Time) event.Event {
	return event.Event{
		ID:          id,
		Title:       "Title " + id,
		Description: "Description " + id,
		StartTime:   start,
		EndTime:     start.Add(1 * time.Hour),
		UserID:      "user-1",
		NotifyAt:    start.Add(-10 * time.Minute),
	}
}

func TestApp_CRUD(t *testing.T) {
	ctx := context.Background()
	a := newTestApp(t)

	// Create
	e := createTestEvent("event-1", time.Now())
	err := a.CreateEvent(ctx, e)
	require.NoError(t, err)

	// Get
	ev, err := a.GetEvent(ctx, "event-1")
	require.NoError(t, err)
	assert.Equal(t, e.Title, ev.Title)

	// Update
	e.Title = "Updated Title"
	err = a.UpdateEvent(ctx, e)
	require.NoError(t, err)

	ev, _ = a.GetEvent(ctx, "event-1")
	assert.Equal(t, "Updated Title", ev.Title)

	// Delete
	err = a.DeleteEvent(ctx, "event-1")
	require.NoError(t, err)

	_, err = a.GetEvent(ctx, "event-1")
	assert.Error(t, err)

	// Delete non-existing
	err = a.DeleteEvent(ctx, "not-found-id")
	assert.Error(t, err)
}

func TestApp_ListEventsByDay(t *testing.T) {
	ctx := context.Background()
	a := newTestApp(t)

	now := time.Date(2025, 7, 22, 10, 0, 0, 0, time.UTC)
	otherDay := now.AddDate(0, 0, 1)

	a.CreateEvent(ctx, createTestEvent("1", now))
	a.CreateEvent(ctx, createTestEvent("2", now.Add(2*time.Hour)))
	a.CreateEvent(ctx, createTestEvent("3", otherDay)) // другой день

	events, err := a.ListEventsByDay(ctx, now)
	require.NoError(t, err)
	assert.Len(t, events, 2)

	for _, e := range events {
		assert.Equal(t, now.Day(), e.StartTime.Day())
	}
}

func TestApp_ListEventsByWeek(t *testing.T) {
	ctx := context.Background()
	a := newTestApp(t)

	ref := time.Date(2025, 7, 22, 10, 0, 0, 0, time.UTC)
	startOfWeek := ref.AddDate(0, 0, -int(ref.Weekday())+1)

	a.CreateEvent(ctx, createTestEvent("1", startOfWeek))
	a.CreateEvent(ctx, createTestEvent("2", startOfWeek.AddDate(0, 0, 3)))
	a.CreateEvent(ctx, createTestEvent("3", startOfWeek.AddDate(0, 0, 7))) // след. неделя

	events, err := a.ListEventsByWeek(ctx, ref)
	require.NoError(t, err)
	assert.Len(t, events, 2)

	for _, e := range events {
		assert.WithinDuration(t, startOfWeek, e.StartTime, 7*24*time.Hour)
	}
}

func TestApp_ListEventsByMonth(t *testing.T) {
	ctx := context.Background()
	a := newTestApp(t)

	ref := time.Date(2025, 7, 22, 10, 0, 0, 0, time.UTC)
	nextMonth := ref.AddDate(0, 1, 0)

	a.CreateEvent(ctx, createTestEvent("1", ref))
	a.CreateEvent(ctx, createTestEvent("2", ref.AddDate(0, 0, 5)))
	a.CreateEvent(ctx, createTestEvent("3", nextMonth)) // другой месяц

	events, err := a.ListEventsByMonth(ctx, ref)
	require.NoError(t, err)
	assert.Len(t, events, 2)

	for _, e := range events {
		assert.Equal(t, ref.Month(), e.StartTime.Month())
	}
}
