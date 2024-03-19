package plugin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/grafana/grafana-app-sdk/plugin"
	"github.com/grafana/grafana-app-sdk/plugin/router"
	"github.com/grafana/grafana-app-sdk/resource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"

	config "github.com/ananasovich/app-o11y-config-manager/pkg/generated/resource/config/v1"
)

type ConfigService interface {
	List(ctx context.Context, namespace string, filters ...string) (*resource.TypedStoreList[*config.Object], error)
	Get(ctx context.Context, id resource.Identifier) (*config.Object, error)
	Add(ctx context.Context, obj *config.Object) (*config.Object, error)
	Update(ctx context.Context, id resource.Identifier, obj *config.Object) (*config.Object, error)
	Delete(ctx context.Context, id resource.Identifier) error
}

func (p *Plugin) handleConfigList(ctx context.Context, req router.JSONRequest) (router.JSONResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "config-list")
	defer span.End()
	filtersRaw := req.URL.Query().Get("filters")
	filters := make([]string, 0)
	if len(filtersRaw) > 0 {
		filters = strings.Split(filtersRaw, ",")
	}
	svc, err := p.service.GetConfigService(ctx)
	if err != nil {
		log.DefaultLogger.With("traceID", span.SpanContext().TraceID()).Error("Error getting ConfigService: "+err.Error(), "error", err)
		return nil, plugin.NewError(http.StatusInternalServerError, err.Error())
	}
	return svc.List(ctx, p.namespace, filters...)
}

func (p *Plugin) handleConfigGet(ctx context.Context, req router.JSONRequest) (router.JSONResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "config-get")
	defer span.End()
	svc, err := p.service.GetConfigService(ctx)
	if err != nil {
		log.DefaultLogger.With("traceID", span.SpanContext().TraceID()).Error("Error getting ConfigService: "+err.Error(), "error", err)
		return nil, plugin.NewError(http.StatusInternalServerError, err.Error())
	}
	obj, err := svc.Get(ctx, resource.Identifier{
		Namespace: p.namespace,
		Name:      req.Vars.MustGet("name"),
	})
	if err != nil {
		if e, ok := err.(errWithStatusCode); ok {
			return nil, plugin.NewError(e.StatusCode(), e.Error())
		} else {
			log.DefaultLogger.With("traceID", span.SpanContext().TraceID()).Error("Error getting Config '"+req.Vars.MustGet("name")+"': "+err.Error(), "error", err)
		}
	}
	return obj, err
}

func (p *Plugin) handleConfigCreate(ctx context.Context, req router.JSONRequest) (router.JSONResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "config-create")
	defer span.End()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, plugin.NewError(http.StatusBadRequest, err.Error())
	}

	t := config.Object{}
	// TODO: this should eventually be unmarshalled via a method in the Object itself, so Thema can handle it
	err = json.Unmarshal(body, &t)
	if err != nil {
		return nil, plugin.NewError(http.StatusBadRequest, err.Error())
	}

	svc, err := p.service.GetConfigService(ctx)
	if err != nil {
		log.DefaultLogger.Error("Error getting ConfigService: " + err.Error())
		return nil, plugin.NewError(http.StatusInternalServerError, err.Error())
	}
	t.StaticMeta.Namespace = p.namespace
	obj, err := svc.Add(ctx, &t)
	if err != nil {
		if e, ok := err.(errWithStatusCode); ok {
			return nil, plugin.NewError(e.StatusCode(), e.Error())
		} else {
			log.DefaultLogger.With("traceID", span.SpanContext().TraceID()).Error("Error creating new Config: "+err.Error(), "error", err)
		}
	}
	return obj, err
}

func (p *Plugin) handleConfigUpdate(ctx context.Context, req router.JSONRequest) (router.JSONResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "config-update")
	defer span.End()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, plugin.NewError(http.StatusBadRequest, err.Error())
	}

	t := config.Object{}
	// TODO: this should eventually be unmarshalled via a method in the Object itself, so Thema can handle it
	err = json.Unmarshal(body, &t)
	if err != nil {
		return nil, plugin.NewError(http.StatusBadRequest, err.Error())
	}

	svc, err := p.service.GetConfigService(ctx)
	if err != nil {
		log.DefaultLogger.With("traceID", span.SpanContext().TraceID()).Error("Error getting ConfigService: "+err.Error(), "error", err)
		return nil, plugin.NewError(http.StatusInternalServerError, err.Error())
	}
	obj, err := svc.Update(ctx, resource.Identifier{
		Namespace: p.namespace,
		Name:      req.Vars.MustGet("name"),
	}, &t)
	if err != nil {
		if e, ok := err.(errWithStatusCode); ok {
			return nil, plugin.NewError(e.StatusCode(), e.Error())
		} else {
			log.DefaultLogger.With("traceID", span.SpanContext().TraceID()).Error("Error updating Config '"+req.Vars.MustGet("name")+"': "+err.Error(), "error", err)
		}
	}
	return obj, err
}

func (p *Plugin) handleConfigDelete(ctx context.Context, req router.JSONRequest) (router.JSONResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "config-delete")
	defer span.End()
	svc, err := p.service.GetConfigService(ctx)
	if err != nil {
		log.DefaultLogger.With("traceID", span.SpanContext().TraceID()).Error("Error getting ConfigService: "+err.Error(), "error", err)
		return nil, plugin.NewError(http.StatusInternalServerError, err.Error())
	}
	err = svc.Delete(ctx, resource.Identifier{
		Namespace: p.namespace,
		Name:      req.Vars.MustGet("name"),
	})
	if err != nil {
		if e, ok := err.(errWithStatusCode); ok {
			return nil, plugin.NewError(e.StatusCode(), e.Error())
		} else {
			log.DefaultLogger.With("traceID", span.SpanContext().TraceID()).Error("Error deleting Config '"+req.Vars.MustGet("name")+"': "+err.Error(), "error", err)
		}
	}
	return nil, err
}
