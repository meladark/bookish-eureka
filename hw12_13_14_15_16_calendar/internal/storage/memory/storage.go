package memorystorage

import (
	"context"
	"sync"
	"time"

	event "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]event.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]event.Event),
	}
}

func (s *Storage) Close(_ context.Context) error {
	return nil
}

func (s *Storage) CreateEvent(_ context.Context, e event.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[e.ID]; exists {
		return event.ErrAlreadyExists
	}

	s.events[e.ID] = e
	return nil
}

func (s *Storage) GetEvent(_ context.Context, id string) (*event.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.events[id]
	if !ok {
		return nil, event.ErrNotFound
	}
	return &e, nil
}

func (s *Storage) UpdateEvent(_ context.Context, e event.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[e.ID]; !exists {
		return event.ErrNotFound
	}

	s.events[e.ID] = e
	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return event.ErrNotFound
	}

	delete(s.events, id)
	return nil
}

func (s *Storage) ListEvents(_ context.Context) ([]event.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]event.Event, 0, len(s.events))
	for _, ev := range s.events {
		events = append(events, ev)
	}
	return events, nil
}

func (s *Storage) EventsToNotify(_ context.Context, now time.Time) ([]event.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []event.Event
	for _, e := range s.events {
		if !e.NotifyAt.IsZero() && e.NotifyAt.Before(now) || e.NotifyAt.Equal(now) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *Storage) DeleteOldEvents(_ context.Context, before time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, e := range s.events {
		if e.EndTime.Before(before) {
			delete(s.events, id)
		}
	}
	return nil
}
