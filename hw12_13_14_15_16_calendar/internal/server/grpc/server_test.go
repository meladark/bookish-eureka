package internalgrpc_test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	pb "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/pb"
	internalgrpc "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/grpc"
	memorystorage "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func dialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func newTestEventProto(id string, start time.Time) *pb.Event {
	return &pb.Event{
		Id:          id,
		Title:       "Title " + id,
		Description: "Description " + id,
		StartTime:   start.UTC().Unix(),
		EndTime:     start.Add(1 * time.Hour).UTC().Unix(),
		UserId:      "user-1",
		NotifyAt:    start.Add(-10 * time.Minute).UTC().Unix(),
	}
}

func setupGrpcTestServer(t *testing.T) (pb.EventServiceClient, func()) {
	t.Helper()
	logFile := "./test.log"
	log := logger.New("debug", logFile)
	t.Cleanup(func() { os.Remove(logFile) })
	storage := memorystorage.New()
	a := app.New(log, storage)
	srv := internalgrpc.NewServer(a, log)

	lis := bufconn.Listen(bufSize)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := srv.ServeListener(ctx, lis); err != nil {
			t.Logf("grpc server error: %v", err)
		}
	}()
	//nolint:staticcheck
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(dialer(lis)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := pb.NewEventServiceClient(conn)

	return client, func() {
		conn.Close()
		cancel()
	}
}

func TestGrpcServer_CRUD(t *testing.T) {
	ctx := context.Background()
	client, cleanup := setupGrpcTestServer(t)
	defer cleanup()

	e := newTestEventProto("event-1", time.Now())
	_, err := client.CreateEvent(ctx, &pb.CreateEventRequest{Event: e})
	require.NoError(t, err)

	resp, err := client.GetEvent(ctx, &pb.GetEventRequest{Id: "event-1"})
	require.NoError(t, err)
	assert.Equal(t, e.Title, resp.Event.Title)

	e.Title = "Updated Title"
	_, err = client.UpdateEvent(ctx, &pb.UpdateEventRequest{Event: e})
	require.NoError(t, err)

	resp, err = client.GetEvent(ctx, &pb.GetEventRequest{Id: "event-1"})
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", resp.Event.Title)

	_, err = client.DeleteEvent(ctx, &pb.DeleteEventRequest{Id: "event-1"})
	require.NoError(t, err)

	_, err = client.GetEvent(ctx, &pb.GetEventRequest{Id: "event-1"})
	assert.Error(t, err)

	_, err = client.DeleteEvent(ctx, &pb.DeleteEventRequest{Id: "not-found-id"})
	assert.Error(t, err)
}

func TestGrpcServer_ListEventsByDay(t *testing.T) {
	ctx := context.Background()
	client, cleanup := setupGrpcTestServer(t)
	defer cleanup()

	now := time.Date(2025, 7, 22, 10, 0, 0, 0, time.UTC)
	otherDay := now.AddDate(0, 0, 1)

	client.CreateEvent(ctx, &pb.CreateEventRequest{Event: newTestEventProto("1", now)})
	client.CreateEvent(ctx, &pb.CreateEventRequest{Event: newTestEventProto("2", now.Add(2*time.Hour))})
	client.CreateEvent(ctx, &pb.CreateEventRequest{Event: newTestEventProto("3", otherDay)})

	resp, err := client.ListEventsByDay(ctx, &pb.ListEventsRequest{Date: now.Unix()})
	require.NoError(t, err)
	assert.Len(t, resp.Events, 2)
	for _, e := range resp.Events {
		assert.Equal(t, now.Day(), time.Unix(e.StartTime, 0).Day())
	}
}

func TestGrpcServer_ListEventsByWeek(t *testing.T) {
	ctx := context.Background()
	client, cleanup := setupGrpcTestServer(t)
	defer cleanup()

	ref := time.Date(2025, 7, 22, 10, 0, 0, 0, time.UTC)

	startOfWeek := ref
	for startOfWeek.Weekday() != time.Monday {
		startOfWeek = startOfWeek.AddDate(0, 0, -1)
	}
	startOfWeek = startOfWeek.Truncate(24 * time.Hour)
	weekEnd := startOfWeek.AddDate(0, 0, 7)

	client.CreateEvent(ctx, &pb.CreateEventRequest{Event: newTestEventProto("1", startOfWeek)})
	client.CreateEvent(ctx, &pb.CreateEventRequest{Event: newTestEventProto("2", startOfWeek.AddDate(0, 0, 3))})
	client.CreateEvent(ctx, &pb.CreateEventRequest{Event: newTestEventProto("3", weekEnd)})

	resp, err := client.ListEventsByWeek(ctx, &pb.ListEventsRequest{
		Date: ref.Unix(),
	})
	require.NoError(t, err)

	for _, e := range resp.Events {
		if e.Id == "3" {
			eventTime := time.Unix(e.StartTime, 0).UTC()
			t.Errorf("Event 3 should be excluded (time: %v, week end: %v)", eventTime, weekEnd)
		}
	}
	assert.Len(t, resp.Events, 2)
}

func TestGrpcServer_ListEventsByMonth(t *testing.T) {
	ctx := context.Background()
	client, cleanup := setupGrpcTestServer(t)
	defer cleanup()

	ref := time.Date(2025, 7, 22, 10, 0, 0, 0, time.UTC)
	nextMonth := ref.AddDate(0, 1, 0)

	client.CreateEvent(ctx, &pb.CreateEventRequest{Event: newTestEventProto("1", ref)})
	client.CreateEvent(ctx, &pb.CreateEventRequest{Event: newTestEventProto("2", ref.AddDate(0, 0, 5))})
	client.CreateEvent(ctx, &pb.CreateEventRequest{Event: newTestEventProto("3", nextMonth)})

	resp, err := client.ListEventsByMonth(ctx, &pb.ListEventsRequest{Date: ref.Unix()})
	require.NoError(t, err)
	assert.Len(t, resp.Events, 2)
	for _, e := range resp.Events {
		assert.Equal(t, ref.Month(), time.Unix(e.StartTime, 0).Month())
	}
}
