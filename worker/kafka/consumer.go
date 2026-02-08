package kafka

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
)

type MessageHandler func(ctx context.Context, msg *TaskMessage) error

type TaskMessage struct {
	TaskID   string `json:"task_id"`
	TraceID  string `json:"trace_id"`
	FilePath string `json:"file_path"`
}

type Consumer struct {
	consumer sarama.ConsumerGroup
}

func NewConsumer(brokers []string, groupID string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	c, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{consumer: c}, nil
}

type consumerHandler struct {
	fn  MessageHandler
	ctx context.Context
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var taskMsg TaskMessage
		if err := json.Unmarshal(msg.Value, &taskMsg); err != nil {
			continue
		}
		h.fn(h.ctx, &taskMsg)
		session.MarkMessage(msg, "")
	}
	return nil
}

func (c *Consumer) Consume(ctx context.Context, topic string, handler MessageHandler) error {
	h := &consumerHandler{fn: handler, ctx: ctx}
	return c.consumer.Consume(ctx, []string{topic}, h)
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
