package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

type EventPayload struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	NotifyAt    string `json:"notifyAt,omitempty"`
	UserId      string `json:"userId,omitempty"`
}

func mustEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func createEvent(t *testing.T, httpAddr string, ev EventPayload) *http.Response {
	t.Helper()
	b, _ := json.Marshal(ev)
	resp, err := http.Post(httpAddr+"/events", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("post event: %v", err)
	}
	return resp
}

func readBodyString(resp *http.Response) string {
	b, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return string(b)
}

func TestCreateEventAndBusinessErrors(t *testing.T) {
	httpAddr := mustEnv("INTEGRATION_APP_ADDR", "http://app:8888")
	now := time.Now().UTC()
	start := now.Add(1 * time.Minute).Format(time.RFC3339)
	end := now.Add(2 * time.Minute).Format(time.RFC3339)
	pgDSN := mustEnv("INTEGRATION_PG_DSN", "postgres://calendar:calendar@postgres:5432/calendar?sslmode=disable")
	db, err := sql.Open("postgres", pgDSN)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	_, _ = db.Exec("DELETE FROM notifications")
	_, _ = db.Exec("DELETE FROM events")

	ev := EventPayload{
		ID:        "it-create-1",
		Title:     "Create test",
		StartTime: start,
		EndTime:   end,
		UserId:    "tester",
	}
	resp := createEvent(t, httpAddr, ev)
	if resp.StatusCode != 201 {
		body := readBodyString(resp)
		t.Fatalf("expected 201 create, got %d body=%s", resp.StatusCode, body)
	}
	_ = resp.Body.Close()

	resp2 := createEvent(t, httpAddr, ev)
	if resp2.StatusCode == 201 {
		_ = resp2.Body.Close()
		t.Fatalf("duplicate creation passed but must fail")
	}
	_ = resp2.Body.Close()

	bad := `{"id":"it-create-2","startTime":"` + start + `","endTime":"` + end + `","userId":"u"}`
	r, err := http.Post(httpAddr+"/events", "application/json", strings.NewReader(bad))
	if err != nil {
		t.Fatalf("post bad payload: %v", err)
	}
	if r.StatusCode != 201 {
		body := readBodyString(r)
		t.Fatalf("expected 201 for invalid payload, got %d body=%s", r.StatusCode, body)
	}
	_ = r.Body.Close()
}

func TestListEventsDayWeekMonth(t *testing.T) {
	httpAddr := mustEnv("INTEGRATION_APP_ADDR", "http://app:8888")
	pgDSN := mustEnv("INTEGRATION_PG_DSN", "postgres://calendar:calendar@postgres:5432/calendar?sslmode=disable")

	db, err := sql.Open("postgres", pgDSN)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	_, _ = db.Exec("DELETE FROM notifications")
	_, _ = db.Exec("DELETE FROM events")

	now := time.Now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.UTC)
	weekStart := dayStart.Add(48 * time.Hour)
	monthStart := dayStart.Add(240 * time.Hour)

	events := []EventPayload{
		{ID: "it-day", Title: "Day", StartTime: dayStart.Format(time.RFC3339), EndTime: dayStart.Add(1 * time.Hour).Format(time.RFC3339), UserId: "u"},
		{ID: "it-week", Title: "Week", StartTime: weekStart.Format(time.RFC3339), EndTime: weekStart.Add(1 * time.Hour).Format(time.RFC3339), UserId: "u"},
		{ID: "it-month", Title: "Month", StartTime: monthStart.Format(time.RFC3339), EndTime: monthStart.Add(1 * time.Hour).Format(time.RFC3339), UserId: "u"},
	}

	for _, ev := range events {
		resp := createEvent(t, httpAddr, ev)
		if resp.StatusCode != 201 {
			body := readBodyString(resp)
			t.Fatalf("create %s failed: %d body=%s", ev.ID, resp.StatusCode, body)
		}
		_ = resp.Body.Close()
	}

	getAndContains := func(path, want string) (bool, string, error) {
		resp, err := http.Get(httpAddr + path)
		if err != nil {
			return false, "", err
		}
		defer resp.Body.Close()
		body := readBodyString(resp)
		return strings.Contains(body, want), body, nil
	}

	contains, body, err := getAndContains(fmt.Sprintf("/events/day?date=%s", dayStart.Format(time.RFC3339)), "it-day")
	if err != nil {
		t.Fatalf("get day: %v", err)
	}
	if !contains || strings.Contains(body, "it-week") {
		t.Fatalf("day listing incorrect; body=%s", body)
	}

	containsDay, _, _ := getAndContains(fmt.Sprintf("/events/week?date=%s", dayStart.Format(time.RFC3339)), "it-day")
	containsWeek, bodyWeek, _ := getAndContains(fmt.Sprintf("/events/week?date=%s", dayStart.Format(time.RFC3339)), "it-week")
	if !containsDay || !containsWeek {
		t.Fatalf("week listing incorrect; body=%s", bodyWeek)
	}

	containsMonthDay, _, _ := getAndContains(fmt.Sprintf("/events/month?date=%s", dayStart.Format(time.RFC3339)), "it-day")
	containsMonthWeek, _, _ := getAndContains(fmt.Sprintf("/events/month?date=%s", dayStart.Format(time.RFC3339)), "it-week")
	containsMonthMonth, bodyMonth, _ := getAndContains(fmt.Sprintf("/events/month?date=%s", dayStart.Format(time.RFC3339)), "it-month")
	if !containsMonthDay || !containsMonthWeek || !containsMonthMonth {
		t.Fatalf("month listing incorrect; body=%s", bodyMonth)
	}
}
