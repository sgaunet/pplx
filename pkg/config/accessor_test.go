package config

import (
	"errors"
	"strings"
	"testing"

	"github.com/sgaunet/pplx/pkg/clerrors"
)

// TestGetValue tests GetValue for all supported field kinds and error paths.
func TestGetValue(t *testing.T) {
	t.Parallel()

	t.Run("string field", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()
		cfg.Defaults.Model = "sonar"

		got, err := GetValue(cfg, "defaults.model")
		if err != nil {
			t.Fatalf("GetValue(defaults.model) unexpected error: %v", err)
		}

		s, ok := got.(string)
		if !ok {
			t.Fatalf("GetValue(defaults.model) returned %T, want string", got)
		}

		if s != "sonar" {
			t.Errorf("GetValue(defaults.model) = %q, want %q", s, "sonar")
		}
	})

	t.Run("float64 field", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()
		cfg.Defaults.Temperature = 0.5

		got, err := GetValue(cfg, "defaults.temperature")
		if err != nil {
			t.Fatalf("GetValue(defaults.temperature) unexpected error: %v", err)
		}

		f, ok := got.(float64)
		if !ok {
			t.Fatalf("GetValue(defaults.temperature) returned %T, want float64", got)
		}

		if f != 0.5 {
			t.Errorf("GetValue(defaults.temperature) = %v, want 0.5", f)
		}
	})

	t.Run("int field", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()
		cfg.Defaults.MaxTokens = 4096

		got, err := GetValue(cfg, "defaults.max_tokens")
		if err != nil {
			t.Fatalf("GetValue(defaults.max_tokens) unexpected error: %v", err)
		}

		n, ok := got.(int)
		if !ok {
			t.Fatalf("GetValue(defaults.max_tokens) returned %T, want int", got)
		}

		if n != 4096 {
			t.Errorf("GetValue(defaults.max_tokens) = %v, want 4096", n)
		}
	})

	t.Run("bool field", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()
		cfg.Output.Stream = true

		got, err := GetValue(cfg, "output.stream")
		if err != nil {
			t.Fatalf("GetValue(output.stream) unexpected error: %v", err)
		}

		b, ok := got.(bool)
		if !ok {
			t.Fatalf("GetValue(output.stream) returned %T, want bool", got)
		}

		if !b {
			t.Errorf("GetValue(output.stream) = false, want true")
		}
	})

	t.Run("string slice field", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()
		cfg.Search.Domains = []string{"example.com", "test.org"}

		got, err := GetValue(cfg, "search.domains")
		if err != nil {
			t.Fatalf("GetValue(search.domains) unexpected error: %v", err)
		}

		domains, ok := got.([]string)
		if !ok {
			t.Fatalf("GetValue(search.domains) returned %T, want []string", got)
		}

		if len(domains) != 2 || domains[0] != "example.com" || domains[1] != "test.org" {
			t.Errorf("GetValue(search.domains) = %v, want [example.com test.org]", domains)
		}
	})

	t.Run("section-level key returns map", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()
		cfg.Defaults.Model = "sonar-pro"

		got, err := GetValue(cfg, "defaults")
		if err != nil {
			t.Fatalf("GetValue(defaults) unexpected error: %v", err)
		}

		m, ok := got.(map[string]any)
		if !ok {
			t.Fatalf("GetValue(defaults) returned %T, want map[string]any", got)
		}

		if len(m) == 0 {
			t.Error("GetValue(defaults) returned empty map")
		}

		if _, hasModel := m["model"]; !hasModel {
			t.Error("GetValue(defaults) map missing 'model' key")
		}
	})

	t.Run("unknown field key returns error with available list", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		_, err := GetValue(cfg, "defaults.nonexistent")
		if err == nil {
			t.Fatal("GetValue(defaults.nonexistent) expected error, got nil")
		}

		if !errors.Is(err, clerrors.ErrOptionNotFound) {
			t.Errorf("GetValue(defaults.nonexistent) error = %v, want ErrOptionNotFound", err)
		}

		if !strings.Contains(err.Error(), "available:") {
			t.Errorf("GetValue(defaults.nonexistent) error %q does not contain 'available:'", err.Error())
		}
	})

	t.Run("unknown section returns error", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		_, err := GetValue(cfg, "unknown.field")
		if err == nil {
			t.Fatal("GetValue(unknown.field) expected error, got nil")
		}

		if !errors.Is(err, clerrors.ErrUnknownSection) {
			t.Errorf("GetValue(unknown.field) error = %v, want ErrUnknownSection", err)
		}
	})
}

// TestSetValue tests SetValue type coercions and error paths.
func TestSetValue(t *testing.T) {
	t.Parallel()

	t.Run("set string", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		if err := SetValue(cfg, "defaults.model", "sonar-pro"); err != nil {
			t.Fatalf("SetValue(defaults.model) unexpected error: %v", err)
		}

		if cfg.Defaults.Model != "sonar-pro" {
			t.Errorf("cfg.Defaults.Model = %q, want %q", cfg.Defaults.Model, "sonar-pro")
		}
	})

	t.Run("set float64", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		if err := SetValue(cfg, "defaults.temperature", "0.7"); err != nil {
			t.Fatalf("SetValue(defaults.temperature) unexpected error: %v", err)
		}

		if cfg.Defaults.Temperature != 0.7 {
			t.Errorf("cfg.Defaults.Temperature = %v, want 0.7", cfg.Defaults.Temperature)
		}
	})

	t.Run("set int", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		if err := SetValue(cfg, "defaults.max_tokens", "8192"); err != nil {
			t.Fatalf("SetValue(defaults.max_tokens) unexpected error: %v", err)
		}

		if cfg.Defaults.MaxTokens != 8192 {
			t.Errorf("cfg.Defaults.MaxTokens = %v, want 8192", cfg.Defaults.MaxTokens)
		}
	})

	t.Run("set bool true", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		if err := SetValue(cfg, "output.stream", "true"); err != nil {
			t.Fatalf("SetValue(output.stream, true) unexpected error: %v", err)
		}

		if !cfg.Output.Stream {
			t.Error("cfg.Output.Stream = false, want true")
		}
	})

	t.Run("set bool false", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()
		cfg.Output.Stream = true

		if err := SetValue(cfg, "output.stream", "false"); err != nil {
			t.Fatalf("SetValue(output.stream, false) unexpected error: %v", err)
		}

		if cfg.Output.Stream {
			t.Error("cfg.Output.Stream = true, want false")
		}
	})

	t.Run("set string slice", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		if err := SetValue(cfg, "search.domains", "a.com,b.com"); err != nil {
			t.Fatalf("SetValue(search.domains) unexpected error: %v", err)
		}

		if len(cfg.Search.Domains) != 2 {
			t.Fatalf("len(cfg.Search.Domains) = %d, want 2", len(cfg.Search.Domains))
		}

		if cfg.Search.Domains[0] != "a.com" || cfg.Search.Domains[1] != "b.com" {
			t.Errorf("cfg.Search.Domains = %v, want [a.com b.com]", cfg.Search.Domains)
		}
	})

	t.Run("set invalid float returns error", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		err := SetValue(cfg, "defaults.temperature", "notanumber")
		if err == nil {
			t.Fatal("SetValue(defaults.temperature, notanumber) expected error, got nil")
		}
	})

	t.Run("set unknown key returns error", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		err := SetValue(cfg, "defaults.nonexistent", "value")
		if err == nil {
			t.Fatal("SetValue(defaults.nonexistent) expected error, got nil")
		}

		if !errors.Is(err, clerrors.ErrOptionNotFound) {
			t.Errorf("SetValue(defaults.nonexistent) error = %v, want ErrOptionNotFound", err)
		}
	})

	t.Run("set unknown section returns error", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		err := SetValue(cfg, "nosuchsection.field", "value")
		if err == nil {
			t.Fatal("SetValue(nosuchsection.field) expected error, got nil")
		}
	})

	t.Run("set key without section returns error", func(t *testing.T) {
		t.Parallel()

		cfg := NewConfigData()

		err := SetValue(cfg, "model", "sonar")
		if err == nil {
			t.Fatal("SetValue(model) expected error for missing section, got nil")
		}

		if !errors.Is(err, clerrors.ErrOptionNotFound) {
			t.Errorf("SetValue(model) error = %v, want ErrOptionNotFound", err)
		}
	})
}

// TestAllKeys tests AllKeys returns a sorted, non-empty slice containing well-known keys.
func TestAllKeys(t *testing.T) {
	t.Parallel()

	keys := AllKeys()

	t.Run("non-empty slice", func(t *testing.T) {
		t.Parallel()

		if len(keys) == 0 {
			t.Fatal("AllKeys() returned empty slice")
		}
	})

	t.Run("contains known keys", func(t *testing.T) {
		t.Parallel()

		knownKeys := []string{
			"defaults.model",
			"search.recency",
			"output.stream",
			"api.key",
		}

		keySet := make(map[string]bool, len(keys))
		for _, k := range keys {
			keySet[k] = true
		}

		for _, want := range knownKeys {
			if !keySet[want] {
				t.Errorf("AllKeys() missing expected key %q", want)
			}
		}
	})

	t.Run("sorted alphabetically", func(t *testing.T) {
		t.Parallel()

		for i := 1; i < len(keys); i++ {
			if keys[i] < keys[i-1] {
				t.Errorf("AllKeys() not sorted: %q comes before %q", keys[i-1], keys[i])
			}
		}
	})

	t.Run("all keys use dot notation", func(t *testing.T) {
		t.Parallel()

		for _, k := range keys {
			if !strings.Contains(k, ".") {
				t.Errorf("AllKeys() key %q does not use dot notation", k)
			}
		}
	})
}
