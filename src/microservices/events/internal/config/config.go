package config

import (
	"os"
)

type Config struct {
	Environment string
	HttpServer  HttpServer
	Kafka       Kafka
}

type HttpServer struct {
	ListenURL string
}

type Kafka struct {
	// TODO: передать на слайс урлов с брокерами
	URL string
}

func Get() (*Config, error) {
	environment := "production"
	if env := os.Getenv("ENV"); env != "" {
		environment = env
	}

	httpListenURL := "0.0.0.0:8082"
	if port := os.Getenv("PORT"); port != "" {
		httpListenURL = "0.0.0.0:" + port
	}

	kafkaURL := "0.0.0.0:9092"
	if url := os.Getenv("KAFKA_BROKERS"); url != "" {
		kafkaURL = url
	}

	return &Config{
		Environment: environment,
		HttpServer: HttpServer{
			ListenURL: httpListenURL,
		},
		Kafka: Kafka{
			URL: kafkaURL,
		},
	}, nil
}
