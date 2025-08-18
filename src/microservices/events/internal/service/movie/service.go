package service_movie

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/pkg/errors"

	"github.com/icecream78/cinema_abyss/src/microservices/events/internal/dto"
)

type kafkaClient interface {
	Publish(topicName string, in any) (dto.PublishedEventDetails, error)
}

const moviesTopic = "movie-events"

type MovieEvent struct {
	ID          int64
	Title       string
	Action      string
	UserID      int64
	Rating      float64
	Genres      []string
	Description string
}

type Service struct {
	kafkaClient kafkaClient
}

func New(
	kafkaClient kafkaClient,
) (*Service, error) {
	return &Service{
		kafkaClient: kafkaClient,
	}, nil
}

// TODO: причесать эту логику. дикий и кривой костыляка, но для MVP пойдет
func (s *Service) HandleTopicName() string {
	return moviesTopic
}

func (s *Service) PublishMovieUpdate(event MovieEvent) (*dto.PublishedEvent, error) {
	publishDetails, err := s.kafkaClient.Publish(moviesTopic, event)
	if err != nil {
		return nil, errors.Wrap(err, "s.kafkaClient.Publish")
	}

	return &dto.PublishedEvent{
		Details: publishDetails,
		Payload: dto.EventPayload{
			ID:        fmt.Sprintf("%d", event.ID),
			Type:      "movie",
			Timestamp: time.Now(),
			Payload:   map[string]string{},
		},
	}, nil
}

func (s *Service) HandleIncomingMovie(e MovieEvent) error {
	slog.Info("handled movie", slog.Int64("name", e.ID))

	return nil
}
