package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-app-sdk/simple"

	config "github.com/ananasovich/app-o11y-config-manager/pkg/generated/resource/config/v1"
	"github.com/ananasovich/app-o11y-config-manager/pkg/watchers"
)

func main() {
	// Configure the default logger to use slog
	logging.DefaultLogger = logging.NewSLogLogger(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	//Load the config from the environment
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		logging.DefaultLogger.With("error", err).Error("Unable to load config from environment")
		panic(err)
	}

	// Load the kube config
	kubeConfig, err := LoadInClusterConfig()
	if err != nil {
		logging.DefaultLogger.With("error", err).Error("Unable to load kubernetes configuration")
		panic(err)
	}

	runner, err := simple.NewOperator(simple.OperatorConfig{
		Name:       "app-o11y-config-manager-operator",
		KubeConfig: kubeConfig.RestConfig,
		Metrics: simple.MetricsConfig{
			Enabled: true,
		},
		Tracing: simple.TracingConfig{
			Enabled: true,
			OpenTelemetryConfig: simple.OpenTelemetryConfig{
				Host:        cfg.OTelConfig.Host,
				Port:        cfg.OTelConfig.Port,
				ConnType:    simple.OTelConnType(cfg.OTelConfig.ConnType),
				ServiceName: cfg.OTelConfig.ServiceName,
			},
		},
		ErrorHandler: func(ctx context.Context, err error) {
			// FIXME: add your own error handling here
			logging.FromContext(ctx).Error(err.Error())
		},
	})
	if err != nil {
		logging.DefaultLogger.With("error", err).Error("Error creating operator")
		panic(err)
	}

	// Wrap our resource watchers in OpinionatedWatchers, then add them to the controller
	configWatcher, err := watchers.NewConfigWatcher()
	if err != nil {
		logging.DefaultLogger.With("error", err).Error("Unable to create ConfigWatcher")
		panic(err)
	}
	err = runner.WatchKind(config.Schema(), configWatcher, simple.ListWatchOptions{
		Namespace: resource.NamespaceAll,
	})
	if err != nil {
		logging.DefaultLogger.With("error", err).Error("Error adding Config watcher to controller")
		panic(err)
	}

	stopCh := make(chan struct{})

	// Signal channel
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		stopCh <- struct{}{}
	}()

	// Run
	logging.DefaultLogger.Info("Starting operator")
	err = runner.Run(stopCh)
	if err != nil {
		logging.DefaultLogger.With("error", err).Error("Operator exited with error")
		panic(err)
	}
	logging.DefaultLogger.Info("Normal operator exit")

}
