package registry

import (
	"context"
	"fmt"
	"sync"

	"aiops2/collector/internal/model"
)

type Collector interface {
	Name() string
	Init(ctx context.Context, cfg model.PluginConfig) error
	Collect(ctx context.Context, jobID string) (*model.JobMeta, error)
	CollectAll(ctx context.Context) ([]*model.JobMeta, error)
	Health(ctx context.Context) error
}

type Registry struct {
	plugins map[string]Collector
	mu      sync.RWMutex
}

func New() *Registry {
	return &Registry{
		plugins: make(map[string]Collector),
	}
}

func (r *Registry) Register(name string, plugin Collector) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if plugin == nil {
		return fmt.Errorf("plugin cannot be nil")
	}
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

func (r *Registry) Get(name string) (Collector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	return plugin, nil
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

func (r *Registry) InitAll(ctx context.Context, configs map[string]model.PluginConfig) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, plugin := range r.plugins {
		cfg, ok := configs[name]
		if !ok {
			cfg = model.PluginConfig{Enabled: true}
		}
		if !cfg.Enabled {
			continue
		}

		if err := plugin.Init(ctx, cfg); err != nil {
			return fmt.Errorf("failed to init plugin %s: %w", name, err)
		}
	}
	return nil
}

func (r *Registry) CollectAll(ctx context.Context) ([]*model.JobMeta, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allJobs []*model.JobMeta
	for name, plugin := range r.plugins {
		jobs, err := plugin.CollectAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("plugin %s collect failed: %w", name, err)
		}
		allJobs = append(allJobs, jobs...)
	}
	return allJobs, nil
}
