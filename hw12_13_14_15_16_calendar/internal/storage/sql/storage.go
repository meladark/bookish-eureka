package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	event "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	//nolint:depguard
	"github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	const checkTableQuery = `
    SELECT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'events' AND typnamespace = 'public'::regnamespace
    ) OR EXISTS (
        SELECT 1 FROM pg_tables WHERE tablename = 'events' AND schemaname = 'public'
    )
`
	var exists bool
	if err := db.QueryRowContext(ctx, checkTableQuery).Scan(&exists); err != nil {
		return nil, err
	}
	//nolint:nestif
	if !exists {
		const createTableQuery = `
			CREATE TABLE IF NOT EXISTS events (
				id TEXT PRIMARY KEY,
				title TEXT NOT NULL,
				description TEXT,
				start_time TIMESTAMP NOT NULL,
				end_time TIMESTAMP NOT NULL,
				user_id TEXT NOT NULL,
				notify_at TIMESTAMP
			)
		`
		if _, err := db.ExecContext(ctx, createTableQuery); err != nil {
			return nil, err
		}
		const createNotifications = `
			CREATE TABLE IF NOT EXISTS notifications (
				id SERIAL PRIMARY KEY,
				event_id TEXT NOT NULL,
				title TEXT NOT NULL,
				scheduled_at TIMESTAMP,
				processed_at TIMESTAMP NOT NULL DEFAULT now()
			)
		`
		if _, err := db.ExecContext(ctx, createNotifications); err != nil {
			return nil, err
		}
	} else {
		const checkColumnsQuery = `
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_name = 'events'
		`

		rows, err := db.QueryContext(ctx, checkColumnsQuery)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		columns := map[string]struct {
			dataType string
			nullable bool
		}{}

		for rows.Next() {
			var colName, dataType, isNullable string
			if err := rows.Scan(&colName, &dataType, &isNullable); err != nil {
				return nil, err
			}
			columns[colName] = struct {
				dataType string
				nullable bool
			}{
				dataType: dataType,
				nullable: isNullable == "YES",
			}
		}

		requiredColumns := map[string]struct {
			dataType string
			nullable bool
		}{
			"id":          {"text", false},
			"title":       {"text", false},
			"description": {"text", true},
			"start_time":  {"timestamp without time zone", false},
			"end_time":    {"timestamp without time zone", false},
			"user_id":     {"text", false},
			"notify_at":   {"timestamp without time zone", true},
		}

		for col, expected := range requiredColumns {
			actual, ok := columns[col]
			if !ok {
				panic("events table schema mismatch: missing column " + col)
			}
			if actual.dataType != expected.dataType || actual.nullable != expected.nullable {
				panic("events table schema mismatch: invalid column " + col)
			}
		}
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close(_ context.Context) error {
	return s.db.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, e event.Event) error {
	query := `
        INSERT INTO events (id, title, description, start_time, end_time, user_id, notify_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := s.db.ExecContext(ctx, query,
		e.ID, e.Title, e.Description, e.StartTime, e.EndTime, e.UserID, e.NotifyAt)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return event.ErrAlreadyExists
			}
		}
		return err
	}
	return nil
}

func (s *Storage) GetEvent(ctx context.Context, id string) (*event.Event, error) {
	query := `
        SELECT id, title, description, start_time, end_time, user_id, notify_at
        FROM events
        WHERE id = $1
    `
	row := s.db.QueryRowContext(ctx, query, id)

	var e event.Event
	if err := row.Scan(&e.ID, &e.Title, &e.Description, &e.StartTime, &e.EndTime, &e.UserID, &e.NotifyAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, event.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	query := `DELETE FROM events WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return event.ErrNotFound
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, e event.Event) error {
	query := `
		UPDATE events
		SET title = $2, description = $3, start_time = $4, end_time = $5, user_id = $6, notify_at = $7
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query,
		e.ID, e.Title, e.Description, e.StartTime, e.EndTime, e.UserID, e.NotifyAt)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return event.ErrNotFound
	}

	return nil
}

func (s *Storage) ListEvents(ctx context.Context) ([]event.Event, error) {
	query := `SELECT id, title, description, start_time, end_time, user_id, notify_at FROM events`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []event.Event
	for rows.Next() {
		e := event.Event{}
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.StartTime, &e.EndTime, &e.UserID, &e.NotifyAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (s *Storage) EventsToNotify(ctx context.Context, now time.Time) ([]event.Event, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, description, start_time, end_time, user_id, notify_at
		FROM events
		WHERE notify_at <= $1 AND notify_at IS NOT NULL
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []event.Event
	for rows.Next() {
		var e event.Event
		err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.StartTime, &e.EndTime, &e.UserID, &e.NotifyAt)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (s *Storage) DeleteOldEvents(ctx context.Context, before time.Time) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE end_time < $1`, before)
	return err
}
