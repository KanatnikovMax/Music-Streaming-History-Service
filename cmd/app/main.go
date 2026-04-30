package main

import (
	"MusicStreamingHistoryService/internal/config"
	"MusicStreamingHistoryService/internal/consumer"
	"MusicStreamingHistoryService/internal/grpc"
	cassandradb "MusicStreamingHistoryService/internal/repository/cassandra"
	"MusicStreamingHistoryService/internal/service"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	cfg := config.MustLoad()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("starting service")

	if err := cassandradb.EnsureKeyspaceIsCreated(cfg.Cassandra); err != nil {
		logger.Fatal("failed to create keyspace", zap.Error(err))
	}
	logger.Info("keyspace ready")

	if err := cassandradb.RunMigrations(cfg.Cassandra.Hosts, cfg.Cassandra.Keyspace); err != nil {
		logger.Fatal("failed to run migrations", zap.Error(err))
	}
	logger.Info("migrations applied")

	session, err := cassandradb.NewSession(cfg.Cassandra)
	if err != nil {
		logger.Fatal("failed to connect to Cassandra", zap.Error(err))
	}
	defer session.Close()

	logger.Info("connected to Cassandra")

	listeningHistoryRepo := cassandradb.NewListeningHistoryRepository(session)
	listeningHistoryService := service.NewListeningHistoryService(listeningHistoryRepo, logger)
	listeningHistoryHandler := grpc.NewListeningHistoryHandler(listeningHistoryService)
	kafkaConsumer := consumer.NewKafkaConsumer(cfg.Kafka, listeningHistoryService, logger)

	grpcServer := grpc.NewServer(cfg.GRPC.Port, logger, listeningHistoryHandler)

	logger.Info("application initialized",
		zap.String("service_type", fmt.Sprintf("%T", listeningHistoryService)),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := kafkaConsumer.Run(ctx); err != nil {
			logger.Fatal("failed to run kafka consumer", zap.Error(err))
		}
	}()

	go func() {
		if err := grpcServer.Run(); err != nil {
			logger.Fatal("failed to start grpc server", zap.Error(err))
		}
	}()

	logger.Info("application started")

	<-ctx.Done()

	logger.Info("stopping service")

	kafkaConsumer.Close()
	grpcServer.GracefulStop()

	logger.Info("service stopped")
}
