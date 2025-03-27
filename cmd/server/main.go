package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ita-av/booking-service/config"
	"github.com/ita-av/booking-service/internal/auth"

	grpcServer "github.com/ita-av/booking-service/internal/grpc"
	"github.com/ita-av/booking-service/internal/repository"
	"github.com/ita-av/booking-service/internal/service"
	pb "github.com/ita-av/booking-service/pkg/api/proto"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Configure logging
	switch cfg.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	log.Info().
		Str("port", cfg.ServerPort).
		Str("mongo_uri", cfg.MongoURI).
		Str("mongo_db", cfg.MongoDB).
		Msg("Starting booking service")

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}

	// Check the connection
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}
	log.Info().Msg("Connected to MongoDB")

	db := mongoClient.Database(cfg.MongoDB)

	// Create repository
	bookingRepo := repository.NewMongoBookingRepository(db)

	// Create service
	bookingService := service.NewBookingService(bookingRepo)

	// Create gRPC server
	bookingServer := grpcServer.NewBookingServer(bookingService)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.ServerPort))
	if err != nil {
		log.Fatal().Err(err).Str("port", cfg.ServerPort).Msg("Failed to listen")
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(auth.AuthInterceptor),
	)
	pb.RegisterBookingServiceServer(s, bookingServer)

	// Enable reflection for tools like grpcurl
	reflection.Register(s)

	// Start server in a goroutine
	go func() {
		log.Info().Str("port", cfg.ServerPort).Msg("gRPC server listening")
		if err := s.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("Failed to serve")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Stop the gRPC server
	s.GracefulStop()

	// Disconnect from MongoDB
	if err := mongoClient.Disconnect(context.Background()); err != nil {
		log.Error().Err(err).Msg("Error disconnecting from MongoDB")
	}

	log.Info().Msg("Server exited properly")
}
