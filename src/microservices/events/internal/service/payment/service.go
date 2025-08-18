package service_payment

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

const userTopic = "payment-events"

type NewPaymentEvent struct {
	UserID      int64
	PaymentID   int64
	Amount      float64
	Status      string
	PaymentType string
	Timestamp   string
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
	return userTopic
}

func (s *Service) PublisNewPayment(event NewPaymentEvent) (*dto.PublishedEvent, error) {
	publishDetails, err := s.kafkaClient.Publish(userTopic, event)
	if err != nil {
		return nil, errors.Wrap(err, "s.kafkaClient.Publish")
	}

	return &dto.PublishedEvent{
		Details: publishDetails,
		Payload: dto.EventPayload{
			ID:        fmt.Sprintf("%d", event.UserID),
			Type:      "user",
			Timestamp: time.Now(),
			Payload:   map[string]string{},
		},
	}, nil
}

func (s *Service) HandleNewPayment(e NewPaymentEvent) error {
	slog.Info("handled new payment", slog.Int64("user_id", e.UserID), slog.Int64("amount", e.UserID))

	return nil
}
