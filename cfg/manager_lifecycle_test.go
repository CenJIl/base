package cfg

import (
	"sync"
	"sync/atomic"
	"testing"
)

type lifecycleConfig struct {
	Name string `toml:"name"`
}

func resetConfigState(t *testing.T) {
	t.Helper()
	if err := Close(); err != nil {
		t.Fatalf("close config watcher: %v", err)
	}
	initOnce = sync.Once{}
	currentConfig = atomic.Pointer[any]{}
	changeHandlers = nil
	cfgLog = nil
}

func TestGetCfgReturnsNilBeforeInitialization(t *testing.T) {
	resetConfigState(t)
	t.Cleanup(func() { resetConfigState(t) })

	if got := GetCfg[lifecycleConfig](); got != nil {
		t.Fatalf("GetCfg() = %#v, want nil before initialization", got)
	}
}

func TestGetCfgRejectsMismatchedConfigurationType(t *testing.T) {
	resetConfigState(t)
	t.Cleanup(func() { resetConfigState(t) })

	storeConfig(&lifecycleConfig{Name: "base"})
	if !IsInitialized[lifecycleConfig]() {
		t.Fatal("IsInitialized() = false, want true")
	}
	if got := GetCfg[struct{ Port int }](); got != nil {
		t.Fatalf("GetCfg() = %#v, want nil for mismatched type", got)
	}
}
