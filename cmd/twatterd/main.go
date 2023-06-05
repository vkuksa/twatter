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
	"github.com/vkuksa/twatter/internal/rabbitmq"
	"github.com/vkuksa/twatter/internal/storage/cockroachdb"
)

const (
	DefaultShutdownTimeout = 10
)

func main() {
	var address string

	flag.StringVar(&address, "address", ":9876", "HTTP Server Address")
	flag.Parse()

	errC, err := run(address)
	if err != nil {
		log.Fatalf("Couldn't run: %s", err)
	}

	if err := <-errC; err != nil {
		log.Fatalf("Error while running: %s", err)
	}
}

func run(address string) (<-chan error, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("zap.NewProduction %w", err)
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

	// // TODO: think how to pass environments here
	// rmq, err := rabbitmq.NewQueue()
	// if err != nil {
	// 	return nil, fmt.Errorf("rabbitmq.NewQueue %w", err)
	// }

	// TODO: add services initialisation here
	// Prometheus

	dbClient, err := cockroachdb.NewClient()
	if err != nil {
		return nil, fmt.Errorf("cockroachdb.NewClient: %w", err)
	}

	srv, err := newServer(serverConfig{
		Address:     address,
		Middlewares: []func(next http.Handler) http.Handler{logging, middleware.Recoverer},
		Logger:      logger,
		DBClient:    dbClient,
	})
	if err != nil {
		return nil, fmt.Errorf("NewServer: %w", err)
	}

	errC := make(chan error, 1)

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-ctx.Done()

		logger.Info("Shutdown signal received")

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer func() {
			_ = logger.Sync()
			// dbconn.Close()	//TODO: modify for cockroach
			// rmq.Close()
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
	DBClient    *cockroachdb.Client
	Queue       *rabbitmq.Queue
}

func newServer(conf serverConfig) (*http.Server, error) {
	router := chi.NewRouter()

	for _, mw := range conf.Middlewares {
		router.Use(mw)
	}

	//TODO: prometheus handlers?

	svc := livefeed.NewService(conf.Logger, conf.DBClient) //, conf.Queue)

	handlers.NewMessageHandler(svc).Register(router)

	return &http.Server{
		Handler: router,
		Addr:    conf.Address,
	}, nil
}
