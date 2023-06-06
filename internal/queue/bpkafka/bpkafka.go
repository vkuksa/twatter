package bpkafka

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vkuksa/twatter/internal/livefeed"
	msgstor "github.com/vkuksa/twatter/internal/storage"
	"go.uber.org/zap"
)

const (
	topic            = "twatter-messages"
	consumerGroup    = "service-worker"
	defaultKafkaAddr = "localhost:9092"
)

// A wrapper for passing logger to objects of segmentio/kafka-go package
type ZapLoggerWrapper struct {
	logger *zap.Logger
}

func (w ZapLoggerWrapper) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	w.logger.Debug(msg)
}

// Simulates backpressure balancing with a kafka streaming
type Queue struct {
	ctx              context.Context
	logger           *zap.Logger
	storage          msgstor.Storage
	msgAddedNotifier livefeed.EventNotifier

	writer     *kafka.Writer
	numWorkers int
	addr       string
}

// Creates new instance of queue
// Takes ctx, logger, storage, addr (kafka address) and n (number of load-balancing workers)
// If kafka address not specified - uses "localhost:9092"
// If number of workers not specified - uses runtime.NumCPU()
// Returns error if kafka dialing fails
func NewBackpressureQueue(ctx context.Context, l *zap.Logger, s msgstor.Storage, notifier livefeed.EventNotifier, addr string, n int) (*Queue, error) {
	if addr == "" {
		addr = defaultKafkaAddr
	}
	if n <= 0 || n == 0 {
		n = runtime.NumCPU()
	}

	_, err := kafka.DialLeader(ctx, "tcp", addr, topic, 0)
	if err != nil {
		return nil, err
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(addr),
		Topic:        topic,
		Async:        true,
		Logger:       ZapLoggerWrapper{logger: l},
		RequiredAcks: 1,
	}

	return &Queue{ctx: ctx, logger: l, storage: s, addr: addr, numWorkers: n, writer: w, msgAddedNotifier: notifier}, nil
}

// Function starts workers, that process messages
func (s *Queue) Start() {
	s.logger.Debug("Starting up insertion workers", zap.Int("amount", s.numWorkers))
	for i := 0; i < s.numWorkers; i++ {
		go s.insertionWorker()
	}
}

// Worker function, that reads message from kafka stream, inserts into storage and notifies that message was inserted
func (s *Queue) insertionWorker() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:                []string{s.addr},
		Topic:                  topic,
		GroupID:                consumerGroup,
		StartOffset:            kafka.LastOffset,
		PartitionWatchInterval: 1 * time.Second,
		JoinGroupBackoff:       1 * time.Second,
	})
	defer reader.Close()

	for {
		kafkaMsg, err := reader.ReadMessage(s.ctx)
		if err != nil {
			s.logger.Error(err.Error())
			return
		}

		s.logger.Debug("storage.Insert: ", zap.String("value", string(kafkaMsg.Value)))
		storedMsg, err := s.storage.InsertMessage(string(kafkaMsg.Value))
		if err != nil {
			s.logger.Error("Storage insertion failed", zap.Error(err), zap.String("value", string(kafkaMsg.Value)))
		} else {
			s.logger.Debug("msgAdded.Notify: ", zap.String("content", string(storedMsg.Content)))
			s.msgAddedNotifier.Notify(storedMsg)
		}

		select {
		case <-s.ctx.Done():
			break
		default:
		}
	}
}

// Adds message to kafka stream
func (s *Queue) Enqueue(ctx context.Context, content string) {
	_ = s.writer.WriteMessages(ctx, kafka.Message{
		Value: []byte(content),
	})
}

// Shutdown queue
func (s *Queue) Shutdown() {
	s.writer.Close()
}
