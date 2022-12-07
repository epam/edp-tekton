package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	triggersApi "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"

	codebaseApiV1 "github.com/epam/edp-codebase-operator/v2/pkg/apis/edp/v1"
	buildInfo "github.com/epam/edp-common/pkg/config"

	"github.com/epam/edp-tekton/pkg/interceptor"
)

const (
	// httpsPort is the port where the interceptor service listens. Use 8443 as it does not require root privileges.
	httpsPort       = 8443
	readTimeout     = 5 * time.Second
	writeTimeout    = 20 * time.Second
	idleTimeout     = 60 * time.Second
	shutDownTimeout = 5 * time.Second
)

type edpInterceptorHandler struct {
	EDPInterceptor interceptor.EDPInterceptorInterface
	Logger         *zap.SugaredLogger
}

type config struct {
	Namespace       string
	InterceptorName string
}

func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %s", err)
	}

	logger := zapLogger.Sugar()

	logBuildInfo(logger)

	var conf *config

	if conf, err = initEnv(); err != nil {
		logger.Fatalf("failed to init env: %v", err)
	}

	clusterConfig := ctrl.GetConfigOrDie()

	scheme := runtime.NewScheme()
	utilruntime.Must(codebaseApiV1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(triggersApi.AddToScheme(scheme))

	client, err := ctrlClient.New(clusterConfig, ctrlClient.Options{Scheme: scheme})
	if err != nil {
		logger.Fatalf("Failed to get client: %v", err)
	}

	secretService := interceptor.NewSecretService(client)

	ctx := context.Background()

	certData, err := secretService.CreateCertsSecret(ctx, conf.Namespace, conf.InterceptorName)
	if err != nil {
		logger.Fatalf("Failed to create certs secret: %v", err)
	}

	logger.Infof("The secret %s was populated with certs ", interceptor.SecretCertsName)

	if err = secretService.UpdateCABundle(ctx, conf.Namespace, conf.InterceptorName, certData.CaCert); err != nil {
		logger.Fatalf("Failed to update cABundle: %v", err)
	}

	logger.Infof("Interceptor %s caBundle updated successfully", conf.InterceptorName)

	mux := http.NewServeMux()
	mux.Handle(
		"/",
		&edpInterceptorHandler{
			EDPInterceptor: interceptor.NewEDPInterceptor(client, logger),
			Logger:         logger,
		},
	)
	mux.HandleFunc("/ready", readinessHandler)

	tlsData := &tls.Config{
		MinVersion: tls.VersionTLS13,
		GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := tls.X509KeyPair(certData.ServerCert, certData.ServerKey)
			if err != nil {
				return nil, err
			}

			return &cert, nil
		},
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", httpsPort),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      mux,
		TLSConfig:    tlsData,
	}

	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed: %v", err)
		}
	}()

	logger.Infof("Listen and serve on port %d", httpsPort)

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

func logBuildInfo(logger *zap.SugaredLogger) {
	v := buildInfo.Get()

	logger.Info("Starting the EDP interceptor",
		"version", v.Version,
		"git-commit", v.GitCommit,
		"git-tag", v.GitTag,
		"build-date", v.BuildDate,
		"go-version", v.Go,
		"go-client", v.KubectlVersion,
		"platform", v.Platform,
	)
}

func initEnv() (*config, error) {
	namespace, ok := os.LookupEnv("SYSTEM_NAMESPACE")
	if !ok {
		return nil, errors.New("env SYSTEM_NAMESPACE is required")
	}

	interceptorName, ok := os.LookupEnv("INTERCEPTOR_NAME")
	if !ok {
		return nil, errors.New("env INTERCEPTOR_NAME is required")
	}

	return &config{
		Namespace:       namespace,
		InterceptorName: interceptorName,
	}, nil
}
