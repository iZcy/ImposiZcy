package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/sirupsen/logrus"
)

type KafkaConnectionInstance struct {
	ID            string
	Name          string
	Broker        string
	Topic         string
	ConsumerGroup sarama.ConsumerGroup
	Handler       *KafkaHandler
	Ready         chan bool
	Cancel        context.CancelFunc
	Running       bool
	mu            sync.Mutex
}

type KafkaManager struct {
	logger       *logrus.Logger
	kafkaHandler *KafkaHandler
	connRepo     *repositories.KafkaConnectionRepository
	connections  map[string]*KafkaConnectionInstance
	mu           sync.RWMutex
}

func NewKafkaManager(
	logger *logrus.Logger,
	kafkaHandler *KafkaHandler,
	connRepo *repositories.KafkaConnectionRepository,
) *KafkaManager {
	return &KafkaManager{
		logger:       logger,
		kafkaHandler: kafkaHandler,
		connRepo:     connRepo,
		connections:  make(map[string]*KafkaConnectionInstance),
	}
}

func (m *KafkaManager) StartAll(ctx context.Context) error {
	if m.connRepo == nil {
		return nil
	}

	conns, err := m.connRepo.GetEnabled(ctx)
	if err != nil {
		return err
	}

	for _, conn := range conns {
		if err := m.StartConnection(ctx, conn); err != nil {
			m.logger.WithError(err).WithField("id", conn.ID.Hex()).Error("Failed to start connection")
		}
	}
	return nil
}

func (m *KafkaManager) StartConnection(ctx context.Context, conn *repositories.KafkaConnection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if instance, exists := m.connections[conn.ID.Hex()]; exists && instance.Running {
		return nil
	}

	brokers := splitByComma(conn.Broker)
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers configured")
	}
	topics := splitByComma(conn.Topic)
	if len(topics) == 0 {
		return fmt.Errorf("no topics configured")
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = conn.ClientID
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	saramaConfig.Consumer.Offsets.Initial = getOffsetValue(conn.AutoOffset)
	saramaConfig.Consumer.Group.Session.Timeout = 30 * time.Second
	saramaConfig.Consumer.Group.Heartbeat.Interval = 10 * time.Second

	consumerGroup, err := sarama.NewConsumerGroup(brokers, conn.GroupID, saramaConfig)
	if err != nil {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	instance := &KafkaConnectionInstance{
		ID:            conn.ID.Hex(),
		Name:          conn.Name,
		Broker:        conn.Broker,
		Topic:         conn.Topic,
		ConsumerGroup: consumerGroup,
		Handler:       m.kafkaHandler,
		Ready:         make(chan bool),
		Running:       false,
	}

	connCtx, cancel := context.WithCancel(context.Background())
	instance.Cancel = cancel
	m.connections[conn.ID.Hex()] = instance

	go m.runConsumer(connCtx, instance, topics)

	select {
	case <-instance.Ready:
		instance.Running = true
	case <-time.After(30 * time.Second):
		instance.Running = true
	}

	m.logger.WithFields(logrus.Fields{
		"id":      conn.ID.Hex(),
		"name":    conn.Name,
		"brokers": brokers,
		"topics":  topics,
	}).Info("Kafka connection started")
	return nil
}

func (m *KafkaManager) runConsumer(ctx context.Context, instance *KafkaConnectionInstance, topics []string) {
	for {
		select {
		case <-ctx.Done():
			instance.Running = false
			return
		default:
			consumer := &managerConsumerHandler{
				manager:      m,
				connectionID: instance.ID,
				logger:       m.logger,
				handler:      instance.Handler,
				ready:        instance.Ready,
			}
			if err := instance.ConsumerGroup.Consume(ctx, topics, consumer); err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					return
				}
				m.logger.WithError(err).WithField("id", instance.ID).Error("Kafka consumer error")
			}
			if ctx.Err() != nil {
				return
			}
		}
	}
}

func (m *KafkaManager) StopConnection(connectionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopConnectionUnsafe(connectionID)
}

func (m *KafkaManager) stopConnectionUnsafe(connectionID string) error {
	instance, exists := m.connections[connectionID]
	if !exists {
		return nil
	}
	if instance.Cancel != nil {
		instance.Cancel()
	}
	if instance.ConsumerGroup != nil {
		instance.ConsumerGroup.Close()
	}
	instance.Running = false
	delete(m.connections, connectionID)
	return nil
}

func (m *KafkaManager) RestartConnection(ctx context.Context, connectionID string) error {
	if err := m.StopConnection(connectionID); err != nil {
		return err
	}
	conn, err := m.connRepo.GetByID(ctx, connectionID)
	if err != nil {
		return err
	}
	return m.StartConnection(ctx, conn)
}

func (m *KafkaManager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id := range m.connections {
		m.stopConnectionUnsafe(id)
	}
}

func (m *KafkaManager) GetConnectionStatus() []*repositories.KafkaConnectionStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var statuses []*repositories.KafkaConnectionStatus
	for _, instance := range m.connections {
		statuses = append(statuses, &repositories.KafkaConnectionStatus{
			ID:        instance.ID,
			Name:      instance.Name,
			Connected: instance.Running,
			Enabled:   true,
		})
	}
	return statuses
}

func (m *KafkaManager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, instance := range m.connections {
		if instance.Running {
			return true
		}
	}
	return false
}

type managerConsumerHandler struct {
	manager      *KafkaManager
	connectionID string
	logger       *logrus.Logger
	handler      *KafkaHandler
	ready        chan bool
}

func (h *managerConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	select {
	case <-h.ready:
	default:
		close(h.ready)
	}
	return nil
}

func (h *managerConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *managerConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}
			h.handler.ProcessRawMessage(session.Context(), message, h.connectionID)
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func splitByComma(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	return result
}

func getOffsetValue(offset string) int64 {
	switch offset {
	case "earliest":
		return sarama.OffsetOldest
	default:
		return sarama.OffsetNewest
	}
}
