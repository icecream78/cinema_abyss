package main

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	client_kafka "github.com/icecream78/cinema_abyss/src/microservices/events/internal/client/kafka"
	"github.com/icecream78/cinema_abyss/src/microservices/events/internal/config"
	"github.com/icecream78/cinema_abyss/src/microservices/events/internal/handler"
	service_movie "github.com/icecream78/cinema_abyss/src/microservices/events/internal/service/movie"
	service_payment "github.com/icecream78/cinema_abyss/src/microservices/events/internal/service/payment"
	service_user "github.com/icecream78/cinema_abyss/src/microservices/events/internal/service/user"
)

type exitCode int

const (
	succesExitCode exitCode = 0
	errorExitCode  exitCode = 1
)

func main() {
	os.Exit(int(run()))
}

func run() exitCode {
	// Handle interrupt signal
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	appConfig, err := config.Get()
	if err != nil {
		slog.Error("during get config", slog.String("error", err.Error()))
		return errorExitCode
	}

	kafkaClient, err := client_kafka.New(appConfig.Kafka.URL)
	if err != nil {
		slog.Error("Error create kafka producer", slog.String("error", err.Error()))
		return errorExitCode
	}
	defer func() {
		if err := kafkaClient.Close(); err != nil {
			slog.Error("Error closing producer", slog.String("error", err.Error()))
		}
	}()

	e := echo.New()

	if appConfig.Environment != "production" {
		e.Use(middleware.Logger())
	}

	movieService, err := service_movie.New(kafkaClient)
	if err != nil {
		slog.Error("Error create movie service", slog.String("error", err.Error()))
		return errorExitCode
	}

	userService, err := service_user.New(kafkaClient)
	if err != nil {
		slog.Error("Error create user service", slog.String("error", err.Error()))
		return errorExitCode
	}

	paymentService, err := service_payment.New(kafkaClient)
	if err != nil {
		slog.Error("Error create payment service", slog.String("error", err.Error()))
		return errorExitCode
	}

	if err = handler.RegisterRouteHandlers(e, movieService, userService, paymentService); err != nil {
		slog.Error("during register http handlers", slog.String("error", err.Error()))
		return errorExitCode
	}

	// подписка на события о фильмах
	if err := client_kafka.SubscribeForTopic(kafkaClient, movieService.HandleTopicName(), movieService.HandleIncomingMovie, signals); err != nil {
		slog.Error("Error create movie subscription", slog.String("error", err.Error()))
		return errorExitCode
	}
	// подписка на события о пользователях
	if err := client_kafka.SubscribeForTopic(kafkaClient, userService.HandleTopicName(), userService.HandleUpdatedUser, signals); err != nil {
		slog.Error("Error create user subscription", slog.String("error", err.Error()))
		return errorExitCode
	}
	// подписка на события об оплатах
	if err := client_kafka.SubscribeForTopic(kafkaClient, paymentService.HandleTopicName(), paymentService.HandleNewPayment, signals); err != nil {
		slog.Error("Error create payment subscription", slog.String("error", err.Error()))
		return errorExitCode
	}

	e.Logger.Fatal(e.Start(appConfig.HttpServer.ListenURL))

	return succesExitCode
}
