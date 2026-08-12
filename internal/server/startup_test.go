package server

import (
	"testing"
	"time"
)

func TestResolveStartupSettings(t *testing.T) {
	cfg := &Config{}
	timeout, retries := resolveStartupSettings(cfg)
	if timeout != 2*time.Second {
		t.Fatalf("expected default timeout 2s, got %v", timeout)
	}
	if retries != 1 {
		t.Fatalf("expected default retries 1, got %d", retries)
	}

	cfg.StartupTimeout = 5 * time.Second
	cfg.StartupRetries = 2
	timeout, retries = resolveStartupSettings(cfg)
	if timeout != 5*time.Second {
		t.Fatalf("expected custom timeout 5s, got %v", timeout)
	}
	if retries != 2 {
		t.Fatalf("expected custom retries 2, got %d", retries)
	}
}
