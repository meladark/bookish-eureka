// internal/server/grpc/server.go
package internalgrpc

import (
	"context"
	"fmt"
	"net"
	"time"

	app "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/app"
	logger "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	pb "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/pb"
	storage "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedEventServiceServer
	app    *app.App
	logger logger.Logger
	grpc   *grpc.Server
}

func NewServer(a *app.App, logger logger.Logger) *Server {
	interceptor := loggingMiddleware(logger)
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor),
	}
	grpcServer := grpc.NewServer(opts...)
	s := &Server{
		app:    a,
		logger: logger,
		grpc:   grpcServer,
	}
	pb.RegisterEventServiceServer(grpcServer, s)
	reflection.Register(grpcServer)
	return s
}

func (s *Server) Start(ctx context.Context, host, port string) error {
	addr := fmt.Sprintf("%s:%s", host, port)
	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	go func() {
		s.logger.Info("gRPC server started at " + addr)
		if err := s.grpc.Serve(lis); err != nil {
			s.logger.Error("gRPC server error: " + err.Error())
		}
	}()
	<-ctx.Done()
	return s.Stop()
}

func (s *Server) Stop() error {
	s.logger.Info("shutting down gRPC server")
	s.grpc.GracefulStop()
	return nil
}

func (s *Server) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	e := req.GetEvent()
	ev := storage.Event{
		ID:          e.Id,
		Title:       e.Title,
		Description: e.Description,
		StartTime:   time.Unix(e.StartTime, 0),
		EndTime:     time.Unix(e.EndTime, 0),
		UserID:      e.UserId,
		NotifyAt:    time.Unix(e.NotifyAt, 0),
	}
	if err := s.app.CreateEvent(ctx, ev); err != nil {
		return nil, err
	}
	return &pb.CreateEventResponse{Event: e}, nil
}

func (s *Server) GetEvent(ctx context.Context, req *pb.GetEventRequest) (*pb.GetEventResponse, error) {
	event, err := s.app.GetEvent(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetEventResponse{
		Event: &pb.Event{
			Id:          event.ID,
			Title:       event.Title,
			Description: event.Description,
			StartTime:   event.StartTime.Unix(),
			EndTime:     event.EndTime.Unix(),
			UserId:      event.UserID,
			NotifyAt:    event.NotifyAt.Unix(),
		},
	}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	e := req.GetEvent()
	ev := storage.Event{
		ID:          e.Id,
		Title:       e.Title,
		Description: e.Description,
		StartTime:   time.Unix(e.StartTime, 0),
		EndTime:     time.Unix(e.EndTime, 0),
		UserID:      e.UserId,
		NotifyAt:    time.Unix(e.NotifyAt, 0),
	}
	if err := s.app.UpdateEvent(ctx, ev); err != nil {
		return nil, err
	}
	return &pb.UpdateEventResponse{Event: e}, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	if err := s.app.DeleteEvent(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pb.DeleteEventResponse{}, nil
}

func (s *Server) ListEventsByDay(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	events, err := s.app.ListEventsByDay(ctx, time.Unix(req.Date, 0))
	if err != nil {
		return nil, err
	}
	return &pb.ListEventsResponse{Events: toProtoEvents(events)}, nil
}

func (s *Server) ListEventsByWeek(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	events, err := s.app.ListEventsByWeek(ctx, time.Unix(req.Date, 0))
	if err != nil {
		return nil, err
	}
	return &pb.ListEventsResponse{Events: toProtoEvents(events)}, nil
}

func (s *Server) ListEventsByMonth(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	events, err := s.app.ListEventsByMonth(ctx, time.Unix(req.Date, 0))
	if err != nil {
		return nil, err
	}
	return &pb.ListEventsResponse{Events: toProtoEvents(events)}, nil
}

func toProtoEvents(events []storage.Event) []*pb.Event {
	protoEvents := make([]*pb.Event, 0, len(events))
	for _, e := range events {
		protoEvents = append(protoEvents, &pb.Event{
			Id:          e.ID,
			Title:       e.Title,
			Description: e.Description,
			StartTime:   e.StartTime.Unix(),
			EndTime:     e.EndTime.Unix(),
			UserId:      e.UserID,
			NotifyAt:    e.NotifyAt.Unix(),
		})
	}
	return protoEvents
}

func (s *Server) ServeListener(ctx context.Context, lis net.Listener) error {
	go func() {
		<-ctx.Done()
		s.logger.Info("shutting down gRPC server")
		s.grpc.GracefulStop()
	}()
	s.logger.Info("starting gRPC server on custom listener")
	return s.grpc.Serve(lis)
}
