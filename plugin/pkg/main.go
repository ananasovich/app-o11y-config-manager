package main

import (
    "context"
	"fmt"
	"os"
	"strings"

	"github.com/grafana/grafana-app-sdk/k8s"
	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/grafana/grafana-app-sdk/metrics"
	sdkPlugin "github.com/grafana/grafana-app-sdk/plugin"
	"github.com/grafana/grafana-app-sdk/plugin/kubeconfig"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/app"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"

    
    
    config "github.com/ananasovich/app-o11y-config-manager/pkg/generated/resource/config/v1"
    "github.com/ananasovich/app-o11y-config-manager/pkg/plugin"
)

const (
    pluginID = "app-o11y-config-manager"
)

func main() {
	// Set the app-sdk logger to use the plugin-sdk logger
    logger := sdkPlugin.NewLogger(log.DefaultLogger.With("pluginID", pluginID))
	logger.Info("starting plugin", "pluginID", pluginID)
	logging.DefaultLogger = logger

	logger.Info("starting plugin", "pluginID", pluginID)

    // app.Manage handles the app plugin lifecycle
	if err := app.Manage(pluginID, newInstanceFactory(logger), app.ManageOpts{
		TracingOpts: tracing.Opts{
			CustomAttributes: []attribute.KeyValue{
				attribute.String("plugin.id", pluginID),
			},
		},
	}); err != nil {
		logger.Error("failed to initialize instance", "err", err)
		os.Exit(1)
	}

	logger.Info("plugin exited normally", "pluginID", pluginID)
	os.Exit(0)
}

//
// GENERATED EXAMPLE SERVICE CODE
// You may want to write your own PluginService code. This example code simply returns unexported 
// service variables for their associated GetX() methods, requiring the initializer to 
// properly set up the services when creating an instance of PluginService. 
// This code initializes each service to a resource.TypedStore for the appropriate schema. 
//

// PluginService implements plugin.Service
type PluginService struct { 
    configService plugin.ConfigService
}

// GetConfigService returns a ConfigService
func (s *PluginService) GetConfigService(ctx context.Context) (plugin.ConfigService, error) {
    return s.configService, nil
}


// newInstanceFactory returns an app.InstanceFactoryFunc to be used with app.Manage
func newInstanceFactory(logger logging.Logger) app.InstanceFactoryFunc {
	return func(ctx context.Context, settings backend.AppInstanceSettings) (instancemgmt.Instance, error) {
        // Load the kubernetes config from the AppInstanceSettings
		kcfg := kubeconfig.NamespacedConfig{}
		err := kubeconfig.NewLoader().LoadFromSettings(settings, &kcfg)
		if err != nil {
            logger.Error("failed to load kubernetes config from settings", "err", err)
			return nil, err
		}

		// Create our client generator, using kubernetes as a store
		clientGenerator := k8s.NewClientRegistry(kcfg.RestConfig, k8s.ClientConfig{
			MetricsConfig: metrics.Config{
				Namespace: strings.ReplaceAll(pluginID, "-", "_"),
			},
		})
		prometheus.MustRegister(clientGenerator.PrometheusCollectors()...)

        // Create our PluginService, then assign values to the individual Kind service(s)
        svc := PluginService{}

		// Create stores for each Kind
        configStore, err := resource.NewTypedStore[*config.Object](config.Schema(), clientGenerator)
        if err != nil {
            logger.Error("failed to create Config store", "err", err)
            return nil, fmt.Errorf("failed to create Config store: %w", err)
        }
        svc.configService = configStore

        // Create the plugin, which allows for CallResource requests to it as an instancemgmt.Instance
        p, err := plugin.New(kcfg.Namespace, &svc)
        if err != nil {
            logger.Error("failed to create plugin instance", "err", err)
			return nil, fmt.Errorf("failed to create plugin instance: %w", err)
        }

		logger.Info("plugin instance provisioned successfully")
		return p, nil
	}
}
