package memorystorage

import (
	"context"
	"sync"

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

func (s *Storage) Close(ctx context.Context) error {
	return nil
}

func (s *Storage) CreateEvent(ctx context.Context, e event.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[e.ID]; exists {
		return event.ErrAlreadyExists
	}

	s.events[e.ID] = e
	return nil
}

func (s *Storage) GetEvent(ctx context.Context, id string) (*event.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.events[id]
	if !ok {
		return nil, event.ErrNotFound
	}
	return &e, nil
}

func (s *Storage) UpdateEvent(ctx context.Context, e event.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[e.ID]; !exists {
		return event.ErrNotFound
	}

	s.events[e.ID] = e
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return event.ErrNotFound
	}

	delete(s.events, id)
	return nil
}

func (s *Storage) ListEvents(ctx context.Context) ([]event.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]event.Event, 0, len(s.events))
	for _, ev := range s.events {
		events = append(events, ev)
	}
	return events, nil
}
