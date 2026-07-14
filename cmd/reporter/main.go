package main

import (
	"flag"
	"log"

	"github.com/go-logr/logr"
	tektonpipelineApi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlLog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsServer "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	codebaseApi "github.com/epam/edp-codebase-operator/v2/api/v1"
	buildInfo "github.com/epam/edp-common/pkg/config"

	"github.com/epam/edp-tekton/pkg/reporter"
	"github.com/epam/edp-tekton/pkg/reporter/collector"
	"github.com/epam/edp-tekton/pkg/reporter/controller"
	"github.com/epam/edp-tekton/pkg/reporter/formatter"
	"github.com/epam/edp-tekton/pkg/reporter/provider"
)

func main() {
	var (
		metricsAddr          string
		probeAddr            string
		enableLeaderElection bool
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", true,
		"Enable leader election for controller manager.")

	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logger := zap.New(zap.UseFlagOptions(&opts))
	ctrl.SetLogger(logger)
	ctrlLog.SetLogger(logger)

	logStartup(logger.WithName("setup"))

	config, err := reporter.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load reporter config: %v", err)
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(codebaseApi.AddToScheme(scheme))
	utilruntime.Must(tektonpipelineApi.AddToScheme(scheme))

	clusterConfig := ctrl.GetConfigOrDie()

	// The reporter opens exactly one watch: review PipelineRuns, filtered
	// server-side by label. Everything else is read on demand via the
	// uncached API reader.
	cacheOptions := cache.Options{
		ByObject: map[ctrlClient.Object]cache.ByObject{
			&tektonpipelineApi.PipelineRun{}: {
				Label: labels.SelectorFromSet(labels.Set{
					reporter.PipelineTypeLabel: reporter.PipelineTypeReview,
				}),
			},
		},
	}

	if config.Namespace != "" {
		cacheOptions.DefaultNamespaces = map[string]cache.Config{
			config.Namespace: {},
		}
	}

	mgr, err := ctrl.NewManager(clusterConfig, ctrl.Options{
		Scheme:                        scheme,
		Cache:                         cacheOptions,
		Metrics:                       metricsServer.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress:        probeAddr,
		LeaderElection:                enableLeaderElection,
		LeaderElectionID:              "edp-tekton-reporter",
		LeaderElectionNamespace:       config.Namespace,
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}

	typedClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	reconciler := controller.NewPipelineRunReconciler(
		mgr.GetClient(),
		mgr.GetAPIReader(),
		collector.New(mgr.GetAPIReader(), collector.NewPodLogFetcher(typedClient), config.TailLines),
		formatter.New(formatter.PortalLinkBuilder{BaseURL: config.PortalBaseURL}),
		provider.New,
		config,
	)

	if err := reconciler.SetupWithManager(mgr); err != nil {
		log.Fatalf("Failed to set up PipelineRun controller: %v", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatalf("Failed to set up health check: %v", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatalf("Failed to set up ready check: %v", err)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("Failed to start manager: %v", err)
	}
}

func logStartup(logger logr.Logger) {
	v := buildInfo.Get()

	logger.Info("Starting the EDP Tekton reporter",
		"version", v.Version,
		"git-commit", v.GitCommit,
		"git-tag", v.GitTag,
		"build-date", v.BuildDate,
		"go-version", v.Go,
		"platform", v.Platform,
	)
}
