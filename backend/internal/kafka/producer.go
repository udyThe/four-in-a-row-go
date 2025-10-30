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
	// First, create the topic if it doesn't exist
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Create game-events topic with 3 partitions
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             "game-events",
			NumPartitions:     3,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Printf("Topic may already exist (this is OK): %v", err)
	} else {
		log.Println("Created Kafka topic: game-events")
	}

	// Now create the writer
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        "game-events",
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	// Test connection with a ping message
	testCtx, testCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer testCancel()

	pingPayload := []byte(`{"event_type":"ping","timestamp":` + strconv.FormatInt(time.Now().Unix(), 10) + `}`)
	testMsg := kafka.Message{
		Key:   []byte("test"),
		Value: pingPayload,
	}

	if err := writer.WriteMessages(testCtx, testMsg); err != nil {
		return nil, err
	}

	log.Println("Kafka producer connected successfully")

	return &Producer{writer: writer}, nil
}

func (p *Producer) SendMessage(ctx context.Context, topic string, data []byte) error {
	// Note: Don't set Topic in the message since Writer already has it configured
	msg := kafka.Message{
		Value: data,
		Time:  time.Now(),
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
