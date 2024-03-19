package plugin

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"

	"github.com/ananasovich/app-o11y-config-manager/pkg/plugin/secure"
	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/grafana/grafana-app-sdk/plugin/kubeconfig"
	"github.com/grafana/grafana-app-sdk/plugin/router"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type Service interface {
	GetConfigService(context.Context) (ConfigService, error)
}

// Plugin is the backend plugin
type Plugin struct {
	router    *router.JSONRouter
	namespace string
	service   Service
}

// Start has the plugin's router start listening over gRPC, and blocks until an unrecoverable error occurs
func (p *Plugin) Start() error {
	return p.router.ListenAndServe()
}

// CallResource allows Plugin to implement grafana-plugin-sdk-go/backend/instancemgmt.Instance for an App plugin,
// Which allows it to be used with grafana-plugin-sdk-go/backend/app.Manage.
// CallResource downstreams all CallResource requests to the router's handler
func (p *Plugin) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	return p.router.CallResource(ctx, req, sender)
}

func New(namespace string, service Service) (*Plugin, error) {
	p := &Plugin{
		router:    router.NewJSONRouter(),
		namespace: namespace,
		service:   service,
	}

	p.router.Use(
		router.NewTracingMiddleware(otel.GetTracerProvider().Tracer("tracing-middleware")),
		router.NewLoggingMiddleware(logging.DefaultLogger),
		kubeconfig.LoadingMiddleware(),
		router.MiddlewareFunc(secure.Middleware))

	// V1 Routes
	v1Subrouter := p.router.Subroute("v1/")

	// Config subrouter
	configSubrouter := v1Subrouter.Subroute("configs/")
	v1Subrouter.Handle("configs", p.handleConfigList, http.MethodGet)
	v1Subrouter.HandleWithCode("configs", p.handleConfigCreate, http.StatusCreated, http.MethodPost)
	configSubrouter.Handle("{name}", p.handleConfigGet, http.MethodGet)
	configSubrouter.Handle("{name}", p.handleConfigUpdate, http.MethodPut)
	configSubrouter.HandleWithCode("{name}", p.handleConfigDelete, http.StatusNoContent, http.MethodDelete)

	return p, nil
}

type errWithStatusCode interface {
	error
	StatusCode() int
}
