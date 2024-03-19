package watchers

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-app-sdk/logging"
	"github.com/grafana/grafana-app-sdk/operator"
	"github.com/grafana/grafana-app-sdk/resource"
	"go.opentelemetry.io/otel"

	config "github.com/ananasovich/app-o11y-config-manager/pkg/generated/resource/config/v1"
)

var _ operator.ResourceWatcher = &ConfigWatcher{}

type ConfigWatcher struct{}

func NewConfigWatcher() (*ConfigWatcher, error) {
	return &ConfigWatcher{}, nil
}

// Add handles add events for config.Object resources.
func (s *ConfigWatcher) Add(ctx context.Context, rObj resource.Object) error {
	ctx, span := otel.GetTracerProvider().Tracer("watcher").Start(ctx, "watcher-add")
	defer span.End()
	object, ok := rObj.(*config.Object)
	if !ok {
		return fmt.Errorf("provided object is not of type *config.Object (name=%s, namespace=%s, kind=%s)",
			rObj.StaticMetadata().Name, rObj.StaticMetadata().Namespace, rObj.StaticMetadata().Kind)
	}

	// TODO
	logging.FromContext(ctx).Debug("Added resource", "name", object.StaticMetadata().Identifier().Name)
	return nil
}

// Update handles update events for config.Object resources.
func (s *ConfigWatcher) Update(ctx context.Context, rOld resource.Object, rNew resource.Object) error {
	ctx, span := otel.GetTracerProvider().Tracer("watcher").Start(ctx, "watcher-update")
	defer span.End()
	oldObject, ok := rOld.(*config.Object)
	if !ok {
		return fmt.Errorf("provided object is not of type *config.Object (name=%s, namespace=%s, kind=%s)",
			rOld.StaticMetadata().Name, rOld.StaticMetadata().Namespace, rOld.StaticMetadata().Kind)
	}

	_, ok = rNew.(*config.Object)
	if !ok {
		return fmt.Errorf("provided object is not of type *config.Object (name=%s, namespace=%s, kind=%s)",
			rNew.StaticMetadata().Name, rNew.StaticMetadata().Namespace, rNew.StaticMetadata().Kind)
	}

	// TODO
	logging.FromContext(ctx).Debug("Updated resource", "name", oldObject.StaticMetadata().Identifier().Name)
	return nil
}

// Delete handles delete events for config.Object resources.
func (s *ConfigWatcher) Delete(ctx context.Context, rObj resource.Object) error {
	ctx, span := otel.GetTracerProvider().Tracer("watcher").Start(ctx, "watcher-delete")
	defer span.End()
	object, ok := rObj.(*config.Object)
	if !ok {
		return fmt.Errorf("provided object is not of type *config.Object (name=%s, namespace=%s, kind=%s)",
			rObj.StaticMetadata().Name, rObj.StaticMetadata().Namespace, rObj.StaticMetadata().Kind)
	}

	// TODO
	logging.FromContext(ctx).Debug("Deleted resource", "name", object.StaticMetadata().Identifier().Name)
	return nil
}

// Sync is not a standard resource.Watcher function, but is used when wrapping this watcher in an operator.OpinionatedWatcher.
// It handles resources which MAY have been updated during an outage period where the watcher was not able to consume events.
func (s *ConfigWatcher) Sync(ctx context.Context, rObj resource.Object) error {
	ctx, span := otel.GetTracerProvider().Tracer("watcher").Start(ctx, "watcher-sync")
	defer span.End()
	object, ok := rObj.(*config.Object)
	if !ok {
		return fmt.Errorf("provided object is not of type *config.Object (name=%s, namespace=%s, kind=%s)",
			rObj.StaticMetadata().Name, rObj.StaticMetadata().Namespace, rObj.StaticMetadata().Kind)
	}

	// TODO
	logging.FromContext(ctx).Debug("Possible resource update", "name", object.StaticMetadata().Identifier().Name)
	return nil
}
