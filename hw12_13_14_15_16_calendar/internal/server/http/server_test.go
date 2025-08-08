package internalhttp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	app "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	logFile := "./test.log"
	log := logger.New("Debug", logFile)
	t.Cleanup(func() { os.Remove(logFile) })
	memStore := memorystorage.New()
	application := app.New(log, memStore)
	server := internalhttp.NewServer(log, application, "127.0.0.1", "8080")
	ts := httptest.NewServer(server.HTTPHandlerForTest())
	t.Cleanup(ts.Close)
	return ts
}

func TestCreateAndListEventsByDay(t *testing.T) {
	ts := setupTestServer(t)
	createEvent(t, ts.URL, `{
		"id": "event-1",
		"title": "Team meeting",
		"description": "Discuss project",
		"startTime": "2025-07-23T10:00:00Z",
		"endTime": "2025-07-23T11:00:00Z",
		"userId": "user-123",
		"notifyAt": "2025-07-23T09:50:00Z"
	}`)
	events := getEvents(t, ts.URL+"/events/day?date=2025-07-23T00:00:00Z")
	require.Len(t, events, 1)
	fmt.Printf("t: %v\n", events)
	require.Equal(t, "event-1", events[0]["ID"])
	require.Equal(t, "Team meeting", events[0]["Title"])
	require.Equal(t, "Discuss project", events[0]["Description"])
	require.Equal(t, "2025-07-23T10:00:00Z", events[0]["StartTime"])
	require.Equal(t, "2025-07-23T11:00:00Z", events[0]["EndTime"])
	require.Equal(t, "user-123", events[0]["UserID"])
	require.Equal(t, "2025-07-23T09:50:00Z", events[0]["NotifyAt"])
}

func TestEventOutsideDayExcluded(t *testing.T) {
	ts := setupTestServer(t)
	createEvent(t, ts.URL, `{
		"id": "event-2",
		"title": "Out of range",
		"startTime": "2025-07-24T10:00:00Z",
		"endTime": "2025-07-24T11:00:00Z"
	}`)
	events := getEvents(t, ts.URL+"/events/day?date=2025-07-23T00:00:00Z")
	require.Len(t, events, 0)
}

func TestListEventsByWeek(t *testing.T) {
	ts := setupTestServer(t)
	createEvent(t, ts.URL, `{
		"id": "event-week",
		"title": "Weekly sync",
		"startTime": "2025-07-24T14:00:00Z",
		"endTime": "2025-07-24T15:00:00Z"
	}`)
	events := getEvents(t, ts.URL+"/events/week?date=2025-07-22T00:00:00Z")
	require.Len(t, events, 1)
	require.Equal(t, "event-week", events[0]["ID"])
}

func TestEventOutsideWeekExcluded(t *testing.T) {
	ts := setupTestServer(t)
	createEvent(t, ts.URL, `{
		"id": "outside-week",
		"title": "Too late",
		"startTime": "2025-07-29T10:00:00Z",
		"endTime": "2025-07-29T11:00:00Z"
	}`)
	events := getEvents(t, ts.URL+"/events/week?date=2025-07-21T00:00:00Z")
	require.Len(t, events, 0)
}

func TestListEventsByMonth(t *testing.T) {
	ts := setupTestServer(t)
	createEvent(t, ts.URL, `{
		"id": "event-month",
		"title": "Monthly report",
		"startTime": "2025-07-05T09:00:00Z",
		"endTime": "2025-07-05T10:00:00Z"
	}`)
	events := getEvents(t, ts.URL+"/events/month?date=2025-07-01T00:00:00Z")
	require.Len(t, events, 1)
	require.Equal(t, "event-month", events[0]["ID"])
}

func TestInvalidStartTimeRejected(t *testing.T) {
	ts := setupTestServer(t)
	//nolint:noctx // просто тесты
	resp, err := http.Post(ts.URL+"/events", "application/json", bytes.NewBufferString(`{
		"id": "bad-time",
		"title": "Fail",
		"startTime": "not-a-date",
		"endTime": "2025-07-23T11:00:00Z"
	}`))
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	defer resp.Body.Close()
}

func TestGetEventsWithoutDateParam(t *testing.T) {
	ts := setupTestServer(t)
	//nolint:noctx // просто тесты
	resp, err := http.Get(ts.URL + "/events/day")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	defer resp.Body.Close()
}

func TestMethodNotAllowed(t *testing.T) {
	ts := setupTestServer(t)
	//nolint:noctx // просто тесты
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/events", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	defer resp.Body.Close()
}

func createEvent(t *testing.T, baseURL, payload string) {
	t.Helper()
	//nolint:noctx // просто тесты
	resp, err := http.Post(baseURL+"/events", "application/json", bytes.NewBufferString(payload))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()
}

func getEvents(t *testing.T, url string) []map[string]interface{} {
	t.Helper()
	//nolint:noctx // просто тесты
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	var events []map[string]interface{}
	err = json.Unmarshal(body, &events)
	require.NoError(t, err)
	return events
}
