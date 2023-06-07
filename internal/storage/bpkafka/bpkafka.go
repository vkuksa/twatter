package bpkafka

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vkuksa/twatter/internal"
	"github.com/vkuksa/twatter/internal/livefeed"
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

type MessageStore interface {
	InsertMessage(string) (internal.Message, error)

	RetrieveAllMessages() ([]internal.Message, error)
}

// Simulates backpressure balancing with a kafka streaming
type Queue struct {
	ctx              context.Context
	logger           *zap.Logger
	store            MessageStore
	msgAddedNotifier livefeed.EventNotifier

	wg         sync.WaitGroup
	writer     *kafka.Writer
	reader     *kafka.Reader
	numWorkers int
}

// Creates new instance of queue
// Takes ctx, logger, storage, addr (kafka address) and n (number of load-balancing workers)
// If kafka address not specified - uses "localhost:9092"
// If number of workers not specified - uses runtime.NumCPU()
// Returns error if kafka dialing fails
func NewBackpressureQueue(ctx context.Context, l *zap.Logger, s MessageStore, addr string, n int) (*Queue, error) {
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

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{addr},
		Topic:       topic,
		GroupID:     consumerGroup,
		Logger:      ZapLoggerWrapper{logger: l},
		StartOffset: kafka.LastOffset,
		MaxWait:     5 * time.Second,
	})
	return &Queue{ctx: ctx, logger: l, store: s, reader: reader, numWorkers: n, writer: w}, nil
}

func (s *Queue) SetMessageAddedNotifier(en livefeed.EventNotifier) {
	s.msgAddedNotifier = en
}

// Function starts workers, that process messages
func (s *Queue) Start() {
	s.logger.Debug("Starting up insertion workers", zap.Int("amount", s.numWorkers))
	for i := 0; i < s.numWorkers; i++ {
		s.wg.Add(1)
		go s.insertionWorker()
	}
}

// Worker function, that reads message from kafka stream, inserts into storage and notifies that message was inserted
func (s *Queue) insertionWorker() {
	defer s.wg.Done()

	for {
		kafkaMsg, err := s.reader.ReadMessage(s.ctx)
		if err != nil {
			s.logger.Error(err.Error())
			return
		}

		s.logger.Debug("storage.Insert: ", zap.String("value", string(kafkaMsg.Value)), zap.ByteString("stack", getStackPrint()))
		storedMsg, err := s.store.InsertMessage(string(kafkaMsg.Value))
		if err != nil {
			s.logger.Error("Storage insertion failed", zap.Error(err), zap.String("value", string(kafkaMsg.Value)))
		} else {
			s.logger.Debug("msgAdded.Notify: ", zap.String("content", string(storedMsg.Content)), zap.ByteString("stack", getStackPrint()))
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
func (s *Queue) InsertMessage(ctx context.Context, content string) {
	_ = s.writer.WriteMessages(ctx, kafka.Message{
		Value: []byte(content),
	})
}

func (s *Queue) RetrieveAllMessages() ([]internal.Message, error) {
	return s.store.RetrieveAllMessages()
}

// Shutdown queue
func (s *Queue) Shutdown() {
	s.reader.Close()
	s.writer.Close()
	s.wg.Wait()
}

func getStackPrint() []byte {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	return b
}
