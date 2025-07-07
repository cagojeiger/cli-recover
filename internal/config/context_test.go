package config

import (
	"context"
	"testing"
)

func TestConfigContext(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Logger.Level = "debug"
	
	// Test WithConfig
	ctx := WithConfig(context.Background(), cfg)
	if ctx == nil {
		t.Fatal("WithConfig returned nil context")
	}
	
	// Test FromContext with valid config
	retrieved, ok := FromContext(ctx)
	if !ok {
		t.Error("FromContext should return true for context with config")
	}
	if retrieved == nil {
		t.Fatal("FromContext returned nil config")
	}
	if retrieved.Logger.Level != "debug" {
		t.Errorf("Expected logger level 'debug', got %s", retrieved.Logger.Level)
	}
	
	// Test FromContext with empty context
	emptyCtx := context.Background()
	retrieved, ok = FromContext(emptyCtx)
	if ok {
		t.Error("FromContext should return false for context without config")
	}
	if retrieved != nil {
		t.Error("FromContext should return nil for context without config")
	}
	
	// Test FromContext with nil context
	retrieved, ok = FromContext(nil)
	if ok {
		t.Error("FromContext should return false for nil context")
	}
	if retrieved != nil {
		t.Error("FromContext should return nil for nil context")
	}
}

func TestMustFromContext(t *testing.T) {
	cfg := DefaultConfig()
	ctx := WithConfig(context.Background(), cfg)
	
	// Test MustFromContext with valid config
	retrieved := MustFromContext(ctx)
	if retrieved == nil {
		t.Fatal("MustFromContext returned nil config")
	}
	if retrieved != cfg {
		t.Error("MustFromContext returned different config instance")
	}
	
	// Test MustFromContext with empty context (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustFromContext should panic for context without config")
		}
	}()
	
	emptyCtx := context.Background()
	MustFromContext(emptyCtx) // This should panic
}

func TestWithConfigNilContext(t *testing.T) {
	cfg := DefaultConfig()
	
	// Test WithConfig with nil context
	ctx := WithConfig(nil, cfg)
	if ctx == nil {
		t.Fatal("WithConfig should create context when given nil")
	}
	
	// Verify config is stored
	retrieved, ok := FromContext(ctx)
	if !ok || retrieved == nil {
		t.Error("Config should be retrievable from context")
	}
}