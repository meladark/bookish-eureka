package storage

import (
	"context"
	"errors"
	"time"
)

type Event struct {
	ID          string    // уникальный идентификатор
	Title       string    // заголовок события
	Description string    // описание
	StartTime   time.Time // дата и время начала
	EndTime     time.Time // дата и время окончания
	UserID      string    // кто создал / владелец события
	NotifyAt    time.Time // время напоминания
}

type Storage interface {
	CreateEvent(ctx context.Context, e *Event) error
	GetEvent(ctx context.Context, id string) (*Event, error)
	UpdateEvent(ctx context.Context, e *Event) error
	DeleteEvent(ctx context.Context, id string) error
	ListEvents(ctx context.Context) ([]*Event, error)
	Close(ctx context.Context) error
}

var (
	ErrNotFound      = errors.New("event not found")
	ErrAlreadyExists = errors.New("event with this ID already exists")
)
