package internalhttp

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	app "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	logger "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	"github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
)

type Server struct {
	logger logger.Logger
	http   *http.Server
	app    *app.App
}

func NewServer(logger logger.Logger, app *app.App, host string, port string) *Server {
	mux := http.NewServeMux()
	s := &Server{
		logger: logger,
		app:    app,
	}

	mux.HandleFunc("/", helloHandler)
	mux.HandleFunc("/events", s.handleCreateEvent)
	mux.HandleFunc("/events/day", s.handleEventsByDay)
	mux.HandleFunc("/events/week", s.handleEventsByWeek)
	mux.HandleFunc("/events/month", s.handleEventsByMonth)

	s.http = &http.Server{
		Addr:              host + ":" + port,
		Handler:           loggingMiddleware(logger)(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s
}

func (s *Server) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		StartTime   string `json:"startTime"`
		EndTime     string `json:"endTime"`
		UserID      string `json:"userId"`
		NotifyAt    string `json:"notifyAt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	start, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		http.Error(w, "invalid startTime", http.StatusBadRequest)
		return
	}
	end, err := time.Parse(time.RFC3339, input.EndTime)
	if err != nil {
		http.Error(w, "invalid endTime", http.StatusBadRequest)
		return
	}

	var notifyAt time.Time
	if input.NotifyAt != "" {
		notifyAt, err = time.Parse(time.RFC3339, input.NotifyAt)
		if err != nil {
			http.Error(w, "invalid notify_at", http.StatusBadRequest)
			return
		}
	}

	event := storage.Event{
		ID:          input.ID,
		Title:       input.Title,
		Description: input.Description,
		StartTime:   start,
		EndTime:     end,
		UserID:      input.UserID,
		NotifyAt:    notifyAt,
	}

	if err := s.app.CreateEvent(r.Context(), event); err != nil {
		http.Error(w, "failed to create event: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleEventsByDay(w http.ResponseWriter, r *http.Request) {
	s.handleListEvents(w, r, "day")
}

func (s *Server) handleEventsByWeek(w http.ResponseWriter, r *http.Request) {
	s.handleListEvents(w, r, "week")
}

func (s *Server) handleEventsByMonth(w http.ResponseWriter, r *http.Request) {
	s.handleListEvents(w, r, "month")
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request, period string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "missing date parameter", http.StatusBadRequest)
		return
	}
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		http.Error(w, "invalid date format", http.StatusBadRequest)
		return
	}
	var events []storage.Event
	switch period {
	case "day":
		events, err = s.app.ListEventsByDay(r.Context(), date)
	case "week":
		events, err = s.app.ListEventsByWeek(r.Context(), date)
	case "month":
		events, err = s.app.ListEventsByMonth(r.Context(), date)
	}
	if err != nil {
		http.Error(w, "failed to list events: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(events)
}

func (s *Server) Start(ctx context.Context) error {
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", s.http.Addr)
	if err != nil {
		return err
	}
	go func() {
		if err := s.http.Serve(ln); err != nil && err != http.ErrServerClosed {
			s.logger.Error("server error: " + err.Error())
		}
	}()
	s.logger.Info("http server started at " + s.http.Addr)
	<-ctx.Done()
	return s.Stop(context.Background())
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down http server")
	return s.http.Shutdown(ctx)
}

func helloHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, Calendar!\n"))
}

func (s *Server) HTTPHandlerForTest() http.Handler {
	return s.http.Handler
}
