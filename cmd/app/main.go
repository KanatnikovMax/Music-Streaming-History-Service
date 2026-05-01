package main

import (
	"MusicStreamingHistoryService/internal/config"
	"MusicStreamingHistoryService/internal/consumer"
	"MusicStreamingHistoryService/internal/grpc"
	log "MusicStreamingHistoryService/internal/logger"
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

	logger := log.MustBuild(cfg.Logger)
	defer logger.Sync()

	logger.Info("starting service", zap.String("env", cfg.Logger.Env))

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

	errCh := make(chan error, 2)

	go func() {
		if err := kafkaConsumer.Run(ctx); err != nil {
			errCh <- err
		}
	}()

	go func() {
		if err := grpcServer.Run(); err != nil {
			errCh <- err
		}
	}()

	logger.Info("application started",
		zap.Int("grpc_port", cfg.GRPC.Port),
		zap.Strings("kafka_brokers", cfg.Kafka.Brokers),
		zap.String("kafka_topic", cfg.Kafka.Topic),
	)

	select {
	case <-ctx.Done():
		logger.Info("gracefully shutting down")
	case err := <-errCh:
		logger.Error("critical error, shutting down", zap.Error(err))
		stop()
	}

	kafkaConsumer.Close()
	grpcServer.GracefulStop()

	logger.Info("service stopped")
}
