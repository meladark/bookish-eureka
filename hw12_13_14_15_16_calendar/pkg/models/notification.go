package models

import "time"

type Notification struct {
	EventID     string    `json:"event_id"`
	Title       string    `json:"title"`
	ScheduledAt time.Time `json:"scheduled_at"`
}
