package main

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/go-multierror"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"testing"
	"time"
)

func newWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func newReader(brokers []string, topic string) *kafka.Reader {
	config := kafka.ReaderConfig{
		Brokers:          brokers,
		Topic:            topic,
		GroupID:          "test-client" + topic,
		MaxWait:          1 * time.Second,
		ReadBatchTimeout: 2 * time.Second,
	}
	err := config.Validate()
	if err != nil {
		log.Println(err)
	}
	return kafka.NewReader(config)
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

func write(t *testing.T, writer *kafka.Writer, key string, msg interface{}) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bs, err := json.Marshal(msg)
	require.NoError(t, err)

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: bs,
	})

	require.NoError(t, err, "не удалось записать сообщение")
}
