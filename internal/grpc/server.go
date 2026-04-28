package grpc

import (
	pb "MusicStreamingHistoryService/pkg/proto/listening_history"
	"context"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	grpcServer *grpc.Server
	port       int
	logger     *zap.Logger
}

func NewServer(
	port int,
	logger *zap.Logger,
	handler *ListeningHistoryHandler,
) *Server {
	loggerFn := func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		switch level {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		}
	}

	recoveryFn := func(p any) error {
		logger.Error("recovered from panic", zap.Any("panic", p))
		return status.Errorf(codes.Internal, "internal server error")
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(logging.LoggerFunc(loggerFn)),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(recoveryFn)),
		),
	)

	pb.RegisterListeningHistoryApiServer(grpcServer, handler)

	return &Server{
		grpcServer: grpcServer,
		port:       port,
		logger:     logger,
	}
}

func (s *Server) Run() error {
	address := fmt.Sprintf(":%d", s.port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	s.logger.Info("grpc server starting", zap.String("address", address))

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("grpc server failed: %w", err)
	}

	return nil
}

func (s *Server) GracefulStop() {
	s.logger.Info("grpc server stopping")
	s.grpcServer.GracefulStop()
}
