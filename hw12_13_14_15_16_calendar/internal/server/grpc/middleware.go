package internalgrpc

import (
	"context"
	"fmt"
	"time"

	logger "github.com/fixme_my_friend/hw12_13_14_15_calendar/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func loggingMiddleware(logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		st, _ := status.FromError(err)
		logLine := fmt.Sprintf(
			"%s [%s] gRPC %s %s %v",
			remoteAddrFromContext(ctx),
			start.Format("02/Jan/2006:15:04:05 -0700"),
			info.FullMethod,
			st.Code(),
			time.Since(start),
		)
		logger.Info(logLine)
		return resp, err
	}
}

func remoteAddrFromContext(ctx context.Context) string {
	pr, ok := peer.FromContext(ctx)
	if ok && pr.Addr != nil {
		return pr.Addr.String()
	}
	return "unknown"
}
