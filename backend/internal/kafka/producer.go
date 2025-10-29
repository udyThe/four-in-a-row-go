package kafka

import (
	"context"
	"log"
	"strconv"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        "game-events",
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to write a lightweight JSON ping message to validate connectivity
	// This avoids non-JSON test payloads being consumed by analytics and causing
	// JSON unmarshal errors in consumers that expect structured events.
	pingPayload := []byte(`{"event_type":"ping","timestamp":` + strconv.FormatInt(time.Now().Unix(), 10) + `}`)
	testMsg := kafka.Message{
		Key:   []byte("test"),
		Value: pingPayload,
	}

	if err := writer.WriteMessages(ctx, testMsg); err != nil {
		return nil, err
	}

	log.Println("Kafka producer connected successfully")

	return &Producer{writer: writer}, nil
}

func (p *Producer) SendMessage(ctx context.Context, topic string, data []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Value: data,
		Time:  time.Now(),
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
