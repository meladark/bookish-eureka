package internalhttp

import (
	"context"
	"net"
	"net/http"
)

type Server struct {
	logger Logger
	http   *http.Server
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

func NewServer(logger Logger) *Server {
	mux := http.NewServeMux()
	s := &Server{
		logger: logger,
	}

	mux.HandleFunc("/", helloHandler)

	s.http = &http.Server{
		Addr:    ":8080",
		Handler: loggingMiddleware(logger)(mux),
	}

	return s
}

func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.http.Addr)
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

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, Calendar!\n"))
}
