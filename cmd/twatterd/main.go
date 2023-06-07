package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/vkuksa/twatter/internal/handlers"
	"github.com/vkuksa/twatter/internal/livefeed"
	"github.com/vkuksa/twatter/internal/storage/bpkafka"
	"github.com/vkuksa/twatter/internal/storage/cockroachdb"
)

const (
	DefaultShutdownTimeout = 3 * time.Second
)

var (
	addr    string
	workers int
)

func main() {
	flag.StringVar(&addr, "addr", ":9876", "HTTP Server Address")
	flag.IntVar(&workers, "queue_workers", 4, "Backpressure queue workers amount")
	flag.Parse()

	errC, err := run(addr)
	if err != nil {
		log.Fatalf("Couldn't run: %s", err)
	}

	if err := <-errC; err != nil {
		log.Fatalf("Error while running: %s", err)
	}
}

func run(address string) (<-chan error, error) {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	logger := zap.Must(zap.NewProduction())
	if os.Getenv("APP_ENV") == "development" {
		logger = zap.Must(zap.NewDevelopment())
	}

	logging := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info(r.Method,
				zap.Time("time", time.Now()),
				zap.String("url", r.URL.String()),
			)

			h.ServeHTTP(w, r)
		})
	}

	storage, err := cockroachdb.NewMessageStore()
	if err != nil {
		return nil, fmt.Errorf("cockroachdb.NewClient: %w", err)
	}

	queue, err := bpkafka.NewBackpressureQueue(ctx, logger, storage, os.Getenv("KAFKA_ADDR"), workers)
	if err != nil {
		return nil, fmt.Errorf("bpkafka.NewQueue: %w", err)
	}

	service := livefeed.NewMessageService(ctx, queue)

	srv, err := newServer(serverConfig{
		Address:     address,
		Middlewares: []func(next http.Handler) http.Handler{logging, middleware.Recoverer},
		Logger:      logger,
		Service:     service,
	})
	if err != nil {
		return nil, fmt.Errorf("NewServer: %w", err)
	}

	errC := make(chan error, 1)

	go func() {
		<-ctx.Done()

		logger.Info("Shutdown signal received")

		ctxTimeout, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)

		defer func() {
			queue.Shutdown()
			_ = storage.Close()
			_ = logger.Sync()
			stop()
			cancel()
			close(errC)
		}()

		srv.SetKeepAlivesEnabled(false)

		if err := srv.Shutdown(ctxTimeout); err != nil {
			errC <- err
		}

		logger.Info("Shutdown completed")
	}()

	go func() {
		queue.Start()

		logger.Info("Listening and serving", zap.String("address", address))

		// "ListenAndServe always returns a non-nil error. After Shutdown or Close, the returned error is
		// ErrServerClosed."
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errC <- err
		}
	}()

	return errC, nil
}

type serverConfig struct {
	Address     string
	Middlewares []func(next http.Handler) http.Handler
	Logger      *zap.Logger
	Service     *livefeed.MessageService
}

func newServer(conf serverConfig) (*http.Server, error) {
	router := chi.NewRouter()

	for _, mw := range conf.Middlewares {
		router.Use(mw)
	}

	handlers.NewMessageHandler(conf.Logger, conf.Service).Register(router)

	return &http.Server{
		Handler: router,
		Addr:    conf.Address,
	}, nil
}
