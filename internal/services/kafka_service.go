package services

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"

	"github.com/iZcy/imposizcy/config"
)

type KafkaService struct {
	consumer sarama.ConsumerGroup
	handler  *KafkaHandler
	cfg      *config.KafkaConfig
	logger   *logrus.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewKafkaService(cfg *config.KafkaConfig, logger *logrus.Logger, handler *KafkaHandler) (*KafkaService, error) {
	if !cfg.Enabled || len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka not configured")
	}

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Version = sarama.V2_5_0_0

	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaService{
		consumer: consumer,
		handler:  handler,
		cfg:      cfg,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (s *KafkaService) Start() error {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				if err := s.consumer.Consume(s.ctx, s.cfg.Topics, s.handler); err != nil {
					s.logger.WithError(err).Error("Kafka consumer error")
				}
				if s.ctx.Err() != nil {
					return
				}
			}
		}
	}()

	s.logger.WithFields(logrus.Fields{
		"brokers": s.cfg.Brokers,
		"topics":  s.cfg.Topics,
		"group":   s.cfg.GroupID,
	}).Info("Kafka consumer started")

	return nil
}

func (s *KafkaService) Stop() error {
	s.cancel()
	return s.consumer.Close()
}
