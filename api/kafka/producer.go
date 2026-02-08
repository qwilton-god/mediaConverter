package kafka

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
)

type Producer interface {
	SendTaskMessage(ctx context.Context, topic string, message *TaskMessage) error
	Close() error
}

type TaskMessage struct {
	TaskID   string `json:"task_id"`
	TraceID  string `json:"trace_id"`
	FilePath string `json:"file_path"`
}

type producer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers []string) (Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	p, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &producer{producer: p}, nil
}

func (p *producer) SendTaskMessage(ctx context.Context, topic string, message *TaskMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(message.TaskID),
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}

func (p *producer) Close() error {
	return p.producer.Close()
}
