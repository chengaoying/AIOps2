package registry

import (
	"context"
	"errors"
	"testing"

	"aiops2/collector/internal/model"
)

type mockCollector struct {
	name   string
	initErr error
	data   []*model.JobMeta
}

func (m *mockCollector) Name() string                         { return m.name }
func (m *mockCollector) Init(ctx context.Context, cfg model.PluginConfig) error { return m.initErr }
func (m *mockCollector) Collect(ctx context.Context, jobID string) (*model.JobMeta, error) {
	return nil, nil
}
func (m *mockCollector) CollectAll(ctx context.Context) ([]*model.JobMeta, error) {
	if m.initErr != nil {
		return nil, m.initErr
	}
	return m.data, nil
}
func (m *mockCollector) Health(ctx context.Context) error { return nil }

func TestRegistry_Register(t *testing.T) {
	r := New()

	err := r.Register("yarn", &mockCollector{name: "yarn"})
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	err = r.Register("yarn", &mockCollector{name: "yarn2"})
	if err == nil {
		t.Error("Register() expected duplicate error")
	}

	err = r.Register("hive", &mockCollector{name: "hive"})
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}
}

func TestRegistry_Get(t *testing.T) {
	r := New()
	r.Register("yarn", &mockCollector{name: "yarn"})

	got, err := r.Get("yarn")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if got.Name() != "yarn" {
		t.Errorf("Get() name = %v, want yarn", got.Name())
	}

	_, err = r.Get("nonexistent")
	if err == nil {
		t.Error("Get() expected error for nonexistent plugin")
	}
}

func TestRegistry_List(t *testing.T) {
	r := New()
	r.Register("yarn", &mockCollector{name: "yarn"})
	r.Register("hive", &mockCollector{name: "hive"})

	names := r.List()
	if len(names) != 2 {
		t.Errorf("List() count = %d, want 2", len(names))
	}
}

func TestRegistry_CollectAll(t *testing.T) {
	r := New()
	r.Register("yarn", &mockCollector{name: "yarn", data: []*model.JobMeta{
		{JobID: "job1", Platform: "YARN"},
	}})
	r.Register("hive", &mockCollector{name: "hive", data: []*model.JobMeta{
		{JobID: "job2", Platform: "HIVE"},
	}})

	jobs, err := r.CollectAll(context.Background())
	if err != nil {
		t.Errorf("CollectAll() error = %v", err)
	}
	if len(jobs) != 2 {
		t.Errorf("CollectAll() count = %d, want 2", len(jobs))
	}
}

func TestRegistry_InitAll(t *testing.T) {
	r := New()
	r.Register("yarn", &mockCollector{name: "yarn"})
	r.Register("hive", &mockCollector{name: "hive"})

	configs := map[string]model.PluginConfig{
		"yarn": {Enabled: true},
		"hive": {Enabled: false},
	}

	err := r.InitAll(context.Background(), configs)
	if err != nil {
		t.Errorf("InitAll() error = %v", err)
	}
}

func TestRegistry_InitAll_Errors(t *testing.T) {
	r := New()
	r.Register("bad", &mockCollector{name: "bad", initErr: errors.New("init failed")})

	configs := map[string]model.PluginConfig{
		"bad": {Enabled: true},
	}

	err := r.InitAll(context.Background(), configs)
	if err == nil {
		t.Error("InitAll() expected error")
	}
}

func TestRegistry_CollectAll_MultipleJobs(t *testing.T) {
	r := New()
	r.Register("yarn", &mockCollector{name: "yarn", data: []*model.JobMeta{
		{JobID: "job1", Platform: "YARN"},
		{JobID: "job2", Platform: "YARN"},
		{JobID: "job3", Platform: "YARN"},
	}})

	jobs, err := r.CollectAll(context.Background())
	if err != nil {
		t.Errorf("CollectAll() error = %v", err)
	}
	if len(jobs) != 3 {
		t.Errorf("CollectAll() count = %d, want 3", len(jobs))
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := New()
	r.Register("yarn", &mockCollector{name: "yarn"})

	_, err := r.Get("yarn")
	if err != nil {
		t.Errorf("Get() after Register() error = %v", err)
	}
}

func TestRegistry_Empty(t *testing.T) {
	r := New()

	names := r.List()
	if len(names) != 0 {
		t.Errorf("List() on empty registry = %d, want 0", len(names))
	}
}

func TestRegistry_NilPlugin(t *testing.T) {
	r := New()

	err := r.Register("nil", nil)
	if err == nil {
		t.Error("Register() expected error for nil plugin")
	}
}

func TestRegistry_DoubleRegister(t *testing.T) {
	r := New()
	r.Register("yarn", &mockCollector{name: "yarn"})

	err := r.Register("yarn", &mockCollector{name: "yarn2"})
	if err == nil {
		t.Error("Double register should error")
	}
}
