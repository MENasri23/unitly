package test

import (
	"context"
	"sync"
	"time"
	"unitly/internal/domain"
)

// Compile-time checks to ensure mocks satisfy the interfaces
var (
	_ domain.Cache = (*MockCache)(nil)
)

// MockCache provides a flexible, function-field-based mock for domain.Cache.
type MockCache struct {
	mu sync.Mutex

	CloseFn         func() error
	GetFn           func(ctx context.Context, key string) (string, error)
	SlidingGetFn    func(ctx context.Context, key string, expiration time.Duration) (string, error)
	IncrFn          func(ctx context.Context, key string) (int64, error)
	ExistsKeyFn     func(ctx context.Context, key string) (bool, error)
	SetFn           func(ctx context.Context, key string, val any, exp time.Duration) error
	SetIfNotExistFn func(ctx context.Context, key string, val any, exp time.Duration) (bool, error)
	DeleteFn        func(ctx context.Context, key string) (bool, error)
	DeleteAllFn     func(ctx context.Context, keys ...string) (int64, error)
	SetMapFn        func(ctx context.Context, key string, values map[string]any) error
	GetMapFn        func(ctx context.Context, key string) (map[string]string, error)
	HIncrByFn       func(ctx context.Context, key, field string, increment int64) (int64, error)
	ExpireFn        func(ctx context.Context, key string, expiration time.Duration) (bool, error)
	TxPipelineFn    func() domain.TxPipeline
	RunScriptFn     func(ctx context.Context, script *domain.Script, keys []string, args ...any) (any, error)

	// Call counts for verifying invocations in tests
	Calls map[string]int
}

func NewMockCache() *MockCache {
	return &MockCache{
		Calls: make(map[string]int),
	}
}

func (m *MockCache) trackCall(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Calls == nil {
		m.Calls = make(map[string]int)
	}
	m.Calls[name]++
}

func (m *MockCache) Close() error {
	m.trackCall("Close")
	if m.CloseFn != nil {
		return m.CloseFn()
	}
	return nil
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	m.trackCall("Get")
	if m.GetFn != nil {
		return m.GetFn(ctx, key)
	}
	panic("MockCache.Get was called but not stubbed")
}

func (m *MockCache) SlidingGet(ctx context.Context, key string, expiration time.Duration) (any, error) {
	m.trackCall("SlidingGet")
	if m.SlidingGetFn != nil {
		return m.SlidingGetFn(ctx, key, expiration)
	}
	panic("MockCache.SlidingGet was called but not stubbed")
}

func (m *MockCache) Incr(ctx context.Context, key string) (int64, error) {
	m.trackCall("Incr")
	if m.IncrFn != nil {
		return m.IncrFn(ctx, key)
	}
	panic("MockCache.Incr was called but not stubbed")
}

func (m *MockCache) ExistsKey(ctx context.Context, key string) (bool, error) {
	m.trackCall("ExistsKey")
	if m.ExistsKeyFn != nil {
		return m.ExistsKeyFn(ctx, key)
	}
	panic("MockCache.ExistsKey was called but not stubbed")
}

func (m *MockCache) Set(ctx context.Context, key string, val any, exp time.Duration) error {
	m.trackCall("Set")
	if m.SetFn != nil {
		return m.SetFn(ctx, key, val, exp)
	}
	panic("MockCache.Set was called but not stubbed")
}

func (m *MockCache) SetIfNotExist(ctx context.Context, key string, val any, exp time.Duration) (bool, error) {
	m.trackCall("SetIfNotExist")
	if m.SetIfNotExistFn != nil {
		return m.SetIfNotExistFn(ctx, key, val, exp)
	}
	panic("MockCache.SetIfNotExist was called but not stubbed")
}

func (m *MockCache) Delete(ctx context.Context, key string) (bool, error) {
	m.trackCall("Delete")
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, key)
	}
	panic("MockCache.Delete was called but not stubbed")
}

func (m *MockCache) DeleteAll(ctx context.Context, keys ...string) (int64, error) {
	m.trackCall("DeleteAll")
	if m.DeleteAllFn != nil {
		return m.DeleteAllFn(ctx, keys...)
	}
	panic("MockCache.DeleteAll was called but not stubbed")
}

func (m *MockCache) SetMap(ctx context.Context, key string, values map[string]any) error {
	m.trackCall("SetMap")
	if m.SetMapFn != nil {
		return m.SetMapFn(ctx, key, values)
	}
	panic("MockCache.SetMap was called but not stubbed")
}

func (m *MockCache) GetMap(ctx context.Context, key string) (map[string]string, error) {
	m.trackCall("GetMap")
	if m.GetMapFn != nil {
		return m.GetMapFn(ctx, key)
	}
	panic("MockCache.GetMap was called but not stubbed")
}

func (m *MockCache) HIncrBy(ctx context.Context, key, field string, increment int64) (int64, error) {
	m.trackCall("HIncrBy")
	if m.HIncrByFn != nil {
		return m.HIncrByFn(ctx, key, field, increment)
	}
	panic("MockCache.HIncrBy was called but not stubbed")
}

func (m *MockCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	m.trackCall("Expire")
	if m.ExpireFn != nil {
		return m.ExpireFn(ctx, key, expiration)
	}
	panic("MockCache.Expire was called but not stubbed")
}

func (m *MockCache) TxPipeline() domain.TxPipeline {
	m.trackCall("TxPipeline")
	if m.TxPipelineFn != nil {
		return m.TxPipelineFn()
	}
	return nil
}

func (m *MockCache) RunScript(ctx context.Context, script *domain.Script, keys []string, args ...any) (any, error) {
	m.trackCall("RunScript")
	if m.RunScriptFn != nil {
		return m.RunScriptFn(ctx, script, keys, args...)
	}
	panic("MockCache.RunScript was called but not stubbed")
}

// -----------------------------------------------------------------------------
// MockTxPipeline
// -----------------------------------------------------------------------------

// MockTxPipeline provides a mock for domain.TxPipeline.
// Customize fields as needed based on the methods declared in your TxPipeline interface.
type MockTxPipeline struct {
	mu sync.Mutex

	ExecFn  func(ctx context.Context) ([]any, error)
	CloseFn func() error

	Calls map[string]int
}

func NewMockTxPipeline() *MockTxPipeline {
	return &MockTxPipeline{
		Calls: make(map[string]int),
	}
}

func (p *MockTxPipeline) trackCall(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Calls == nil {
		p.Calls = make(map[string]int)
	}
	p.Calls[name]++
}

func (p *MockTxPipeline) Exec(ctx context.Context) ([]any, error) {
	p.trackCall("Exec")
	if p.ExecFn != nil {
		return p.ExecFn(ctx)
	}
	return nil, nil
}

func (p *MockTxPipeline) Close() error {
	p.trackCall("Close")
	if p.CloseFn != nil {
		return p.CloseFn()
	}
	return nil
}
