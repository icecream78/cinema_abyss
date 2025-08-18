package client_kafka

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"

	"github.com/icecream78/cinema_abyss/src/microservices/events/internal/dto"
)

type Client struct {
	producer sarama.SyncProducer
	consumer sarama.Consumer
}

func New(kafkaURL string) (*Client, error) {
	var producer sarama.SyncProducer
	var consumer sarama.Consumer
	var err error

	// TODO: унести в удобный конфигуратор, чтобы было почище
	maxRetries := 5
	retryDelay := time.Second * 5
	for range maxRetries {
		producer, err = sarama.NewSyncProducer(
			[]string{kafkaURL},
			defaultKafkaConfig(),
		)
		if err != nil {
			time.Sleep(retryDelay)
			continue
			// return nil, errors.Wrap(err, "sarama.NewSyncProducer")
		}

		consumer, err = sarama.NewConsumer(
			[]string{kafkaURL},
			defaultKafkaConfig(),
		)
		if err != nil {
			time.Sleep(retryDelay)
			continue
			// return nil, errors.Wrap(err, "sarama.NewConsumer")
		}

		break
	}
	if err != nil {
		return nil, errors.Wrap(err, "init consumer/producer")
	}

	return &Client{
		producer: producer,
		consumer: consumer,
	}, nil
}

func (c *Client) Close() error {
	if err := c.producer.Close(); err != nil {
		return errors.Wrap(err, "c.producer.Close")
	}

	return nil
}

func (c *Client) Publish(topicName string, in any) (dto.PublishedEventDetails, error) {
	inBytes, err := json.Marshal(in)
	if err != nil {
		return dto.PublishedEventDetails{}, errors.Wrap(err, "json.Marshal")
	}

	msg := sarama.ProducerMessage{
		Topic: topicName,
		Value: sarama.ByteEncoder(inBytes),
	}

	partition, offset, err := c.producer.SendMessage(&msg)
	if err != nil {
		return dto.PublishedEventDetails{}, errors.Wrap(err, "c.producer.SendMessage")
	}

	return dto.PublishedEventDetails{
		Status:    "success",
		Partition: int64(partition),
		Offset:    offset,
	}, nil
}

type SubscribeFunc[T any] func(msg T) error

func SubscribeForTopic[T any](
	c *Client,
	topicName string,
	callbackFn SubscribeFunc[T],
	finishCh chan os.Signal,
) error {
	partitions, err := c.consumer.Partitions(topicName)
	if err != nil {
		return errors.Wrap(err, "c.consumer.Partitions")
	}

	var wg sync.WaitGroup
	for _, partition := range partitions {
		wg.Add(1)

		pc, err := c.consumer.ConsumePartition(topicName, partition, sarama.OffsetNewest)
		if err != nil {
			return errors.Wrap(err, "c.consumer.ConsumePartition")
		}

		go func(pc sarama.PartitionConsumer) {
			defer wg.Done()
			defer pc.AsyncClose()

		ConsumerLoop:
			for {
				select {
				case msg := <-pc.Messages():
					var unmarshalledMsg T
					if err := json.Unmarshal(msg.Value, &unmarshalledMsg); err != nil {
						slog.Error("unmarshal input message", slog.String("error", err.Error()))
					}
					if err := callbackFn(unmarshalledMsg); err != nil {
						slog.Error("handle message error", slog.String("error", err.Error()))
					}
				case err := <-pc.Errors():
					log.Printf("Error: %v", err)
				case <-finishCh:
					break ConsumerLoop
				}
			}
		}(pc)
	}

	return nil
}

func defaultKafkaConfig() *sarama.Config {
	config := sarama.NewConfig()

	config.Version = sarama.V2_7_0_0

	// producer params
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll

	// consumer params
	config.Consumer.Return.Errors = true

	return config
}
