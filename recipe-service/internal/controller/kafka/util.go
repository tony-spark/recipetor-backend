package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	"github.com/segmentio/kafka-go"
)

func logdf(msg string, a ...interface{}) {
	log.Debug().Msgf(msg, a...)
}

func logef(msg string, a ...interface{}) {
	log.Error().Msgf(msg, a...)
}

func newReader(brokers []string, group string, topic string) (*kafka.Reader, error) {
	config := kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     group,
		Logger:      kafka.LoggerFunc(logdf),
		ErrorLogger: kafka.LoggerFunc(logef),
	}
	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid kafka config: %w", err)
	}
	return kafka.NewReader(config), nil
}

func newWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:        kafka.TCP(brokers...),
		Topic:       topic,
		Balancer:    &kafka.LeastBytes{},
		Logger:      kafka.LoggerFunc(logdf),
		ErrorLogger: kafka.LoggerFunc(logef),
	}
}

func readDTO(ctx context.Context, reader *kafka.Reader, obj interface{}) error {
	m, err := reader.ReadMessage(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error receiving message")
		return err
	}

	err = json.Unmarshal(m.Value, obj)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal message")
	}

	return err
}

func write(writer *kafka.Writer, key string, msg interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bs, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal outcoming message")
		return
	}

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: bs,
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to write message")
	}
}

func closeAll(closers ...io.Closer) error {
	var result error
	for _, closer := range closers {
		err := closer.Close()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}