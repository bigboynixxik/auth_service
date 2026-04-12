package grpc

import (
	"auth-service/pkg/logger"
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
)

func LoggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		l := log.With(slog.String("method", info.FullMethod))

		ctx = logger.WithContext(ctx, l)

		start := time.Now()

		resp, err := handler(ctx, req)
		if err != nil {
			l.Error("incoming gRPC request failed",
				slog.String("error", err.Error()),
				slog.Duration("duration", time.Since(start)))
		} else {
			l.Info("incoming gRPC response",
				slog.Duration("duration", time.Since(start)))
		}
		return resp, err
	}
}
