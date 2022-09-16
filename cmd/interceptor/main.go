package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApiV1 "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1"

	"github.com/epam/edp-tekton/pkg/interceptor"
)

const (
	// Port is the port that the port that interceptor service listens on.
	Port            = 8080
	readTimeout     = 5 * time.Second
	writeTimeout    = 20 * time.Second
	idleTimeout     = 60 * time.Second
	shutDownTimeout = 5 * time.Second
)

type edpInterceptorHandler struct {
	EDPInterceptor interceptor.EDPInterceptorInterface
	Logger         *zap.SugaredLogger
}

func main() {
	clusterConfig := ctrl.GetConfigOrDie()

	scheme := runtime.NewScheme()
	utilruntime.Must(codebaseApiV1.AddToScheme(scheme))

	client, err := ctrlClient.New(clusterConfig, ctrlClient.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}

	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %s", err)
	}

	logger := zapLogger.Sugar()

	//TODO: We need to move to https server. See: https://tekton.dev/docs/triggers/clusterinterceptors/#running-clusterinterceptor-as-https
	mux := http.NewServeMux()
	mux.Handle(
		"/",
		&edpInterceptorHandler{
			EDPInterceptor: interceptor.NewEDPInterceptor(client, logger),
			Logger:         logger,
		},
	)
	mux.HandleFunc("/ready", readinessHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", Port),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed: %v", err)
		}
	}()

	logger.Infof("Listen and serve on port %d", Port)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	logger.Info("Server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), shutDownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server shutdown failed: %+v", err)
	}

	logger.Info("Server exited properly")
}

func (h *edpInterceptorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := h.EDPInterceptor.Execute(r)

	if err != nil {
		interceptorErr := &interceptor.HTTPError{}

		if errors.As(err, interceptorErr) {
			h.Logger.Infof("HTTP %d - %s", interceptorErr.Status(), interceptorErr)
			http.Error(w, interceptorErr.Error(), interceptorErr.Status())

			return
		}

		h.Logger.Errorf("Non Status Error: %s", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "application/json")

	if _, err := w.Write(b); err != nil {
		h.Logger.Errorf("Failed to write response: %s", err)
	}
}

func readinessHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
