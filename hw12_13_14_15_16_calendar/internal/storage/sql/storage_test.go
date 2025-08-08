package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	//nolint:depguard
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	//nolint:depguard
	event "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*Storage, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	store := &Storage{db: db}

	// Финализатор для закрытия БД
	cleanup := func() {
		db.Close()
	}

	return store, mock, cleanup
}

func TestCreateEvent(t *testing.T) {
	store, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()
	e := event.Event{
		ID:          "123",
		Title:       "Test Event",
		Description: "Test Description",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Hour),
		UserID:      "Mike",
		NotifyAt:    time.Now().Add(30 * time.Minute),
	}

	mock.ExpectExec("INSERT INTO events").
		WithArgs(e.ID, e.Title, e.Description, e.StartTime, e.EndTime, e.UserID, e.NotifyAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.CreateEvent(ctx, e)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetEvent(t *testing.T) {
	store, mock, cleanup := setupMockDB(t)
	defer cleanup()

	ctx := context.Background()
	id := "123"
	start := time.Now()
	end := start.Add(1 * time.Hour)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "start_time", "end_time", "user_id", "notify_at"}).
		AddRow(id, "Test Event", "Test Description", start, end, "Mike", time.Now().Add(30*time.Minute))

	mock.ExpectQuery(
		"SELECT id, title, description, start_time, end_time, user_id, notify_at FROM events WHERE id = \\$1").
		WithArgs(id).
		WillReturnRows(rows)

	e, err := store.GetEvent(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, e)
	require.Equal(t, id, e.ID)
	require.Equal(t, "Test Event", e.Title)
	require.Equal(t, "Test Description", e.Description)
	require.WithinDuration(t, start, e.StartTime, time.Second)
	require.WithinDuration(t, end, e.EndTime, time.Second)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLStorage_DeleteUpdateList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	store := &Storage{db: db}
	ctx := context.Background()

	event := event.Event{
		ID:          "123",
		Title:       "Test Event",
		Description: "Test Description",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Hour),
		UserID:      "Mike",
		NotifyAt:    time.Now().Add(30 * time.Minute),
	}

	mock.ExpectExec(`DELETE FROM events WHERE id = \$1`).
		WithArgs(event.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.DeleteEvent(ctx, event.ID)
	require.NoError(t, err)

	mock.ExpectExec(`UPDATE events`).
		WithArgs(event.ID, event.Title, event.Description, event.StartTime, event.EndTime, event.UserID, event.NotifyAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = store.UpdateEvent(ctx, event)
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id", "title", "description", "start_time", "end_time", "user_id", "notify_at"}).
		AddRow(event.ID, event.Title, event.Description, event.StartTime, event.EndTime, event.UserID, event.NotifyAt)

	mock.ExpectQuery(`SELECT id, title, description, start_time, end_time, user_id, notify_at FROM events`).
		WillReturnRows(rows)

	events, err := store.ListEvents(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, event.ID, events[0].ID)
	require.Equal(t, event.Title, events[0].Title)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewIntegration(t *testing.T) {
	dsn := "postgres://calendar:calendar@localhost:5499/calendar?sslmode=disable"

	if !isDBAvailable(dsn) {
		t.Skip("Database not available, skipping integration test")
	}
	dropEventsTable(t, dsn)
	s, err := New(dsn)
	require.NoError(t, err)
	require.NotNil(t, s)

	simplEvent := event.Event{
		ID:          "123",
		Title:       "Test Event",
		Description: "Test Description",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Hour),
		UserID:      "Mike",
	}

	if err := s.CreateEvent(context.Background(), simplEvent); err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	err = s.CreateEvent(context.Background(), simplEvent)
	fmt.Printf("t: %v\n", err)
	require.ErrorAs(t, err, &event.ErrAlreadyExists, "CreateEvent should return ErrAlreadyExists")

	if e, err := s.GetEvent(context.Background(), "123"); err != nil {
		require.Equal(t, simplEvent, e)
		if !errors.Is(err, event.ErrNotFound) {
			t.Fatalf("GetEvent failed: %v", err)
		}
	}
	if _, err := s.GetEvent(context.Background(), "124"); err != nil {
		require.ErrorAs(t, err, &event.ErrNotFound, "GetEvent should return ErrNotFound")
	}
	if err := s.DeleteEvent(context.Background(), "123"); err != nil {
		if !errors.Is(err, event.ErrNotFound) {
			t.Fatalf("DeleteEvent failed: %v", err)
		}
	}
	if err := s.DeleteEvent(context.Background(), "124"); err != nil {
		require.ErrorAs(t, err, &event.ErrNotFound, "DeleteEvent should return ErrNotFound")
	}
	require.NoError(t, s.Close(context.Background()))
	s, err = New(dsn)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.NoError(t, s.Close(context.Background()))
	dropEventsTable(t, dsn)
}

func isDBAvailable(dsn string) bool {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return false
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return false
	}

	return true
}

func dropEventsTable(t *testing.T, dsn string) {
	t.Helper()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()
	//nolint:noctx // просто тесты
	_, err = db.Exec(`DROP TABLE IF EXISTS public.events`)
	require.NoError(t, err)
}
