package memorystorage

import (
	"context"
	"testing"
	"time"

	//nolint:depguard
	event "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	ctx := context.Background()
	store := New()

	// Создаем событие
	e := event.Event{
		ID:          "123",
		Title:       "Test Event",
		Description: "Test Description",
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(1 * time.Hour),
		UserID:      "Mike",
		NotifyAt:    time.Now().Add(30 * time.Minute),
	}

	err := store.CreateEvent(ctx, e)
	require.NoError(t, err, "CreateEvent failed")

	err = store.CreateEvent(ctx, e)
	require.ErrorAs(t, err, &event.ErrAlreadyExists, "CreateEvent should return ErrAlreadyExists")

	got, err := store.GetEvent(ctx, e.ID)
	require.NoError(t, err, "GetEvent failed")
	require.NotNil(t, got)
	require.Equal(t, e.Title, got.Title)

	e.Title = "Updated Title"
	err = store.UpdateEvent(ctx, e)
	require.NoError(t, err, "UpdateEvent failed")

	got, err = store.GetEvent(ctx, e.ID)
	require.NoError(t, err)
	require.Equal(t, "Updated Title", got.Title)

	events, err := store.ListEvents(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)

	err = store.DeleteEvent(ctx, e.ID)
	require.NoError(t, err)

	got, err = store.GetEvent(ctx, e.ID)
	require.ErrorAs(t, err, &event.ErrNotFound, "GetEvent should return ErrNotFound")
	require.Nil(t, got)

	events, err = store.ListEvents(ctx)
	require.NoError(t, err)
	require.Len(t, events, 0)

	err = store.DeleteEvent(ctx, e.ID)
	require.ErrorAs(t, err, &event.ErrNotFound, "DeleteEvent should return ErrNotFound")

	e.Title = "Updated Title"
	err = store.UpdateEvent(ctx, e)
	require.ErrorAs(t, err, &event.ErrNotFound, "UpdateEvent should return ErrNotFound")

	store.Close(ctx)
}
