package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sgaunet/pplx/pkg/config"
)

// BenchmarkConfigLoad benchmarks config file loading performance.
func BenchmarkConfigLoad(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create a test config
	fixtureData, err := os.ReadFile(filepath.Join("testdata", "valid_config.yaml"))
	if err != nil {
		b.Fatalf("Failed to read fixture: %v", err)
	}
	if err := os.WriteFile(configPath, fixtureData, configFilePermission); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := config.NewLoader()
		if err := loader.LoadFrom(configPath); err != nil {
			b.Fatalf("LoadFrom failed: %v", err)
		}
	}
}

// BenchmarkConfigValidation benchmarks config validation performance.
func BenchmarkConfigValidation(b *testing.B) {
	cfg := config.NewConfigData()
	cfg.Defaults.Model = "sonar"
	cfg.Defaults.Temperature = 0.7
	cfg.Defaults.MaxTokens = 4096
	cfg.Search.Mode = "web"
	cfg.Search.ContextSize = "medium"
	cfg.Output.Stream = false

	validator := config.NewValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validator.Validate(cfg); err != nil {
			b.Fatalf("Validate failed: %v", err)
		}
	}
}

// BenchmarkTemplateLoading benchmarks template loading performance.
func BenchmarkTemplateLoading(b *testing.B) {
	templates := []struct {
		name string
	}{
		{name: config.TemplateResearch},
		{name: config.TemplateCreative},
		{name: config.TemplateNews},
		{name: config.TemplateFullExample},
	}

	for _, tt := range templates {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cfg, err := config.LoadTemplate(tt.name)
				if err != nil {
					b.Fatalf("LoadTemplate failed: %v", err)
				}
				if cfg == nil {
					b.Fatal("LoadTemplate returned nil")
				}
			}
		})
	}
}

// BenchmarkAnnotatedConfigGeneration benchmarks annotated config generation.
func BenchmarkAnnotatedConfigGeneration(b *testing.B) {
	cfg := config.NewConfigData()
	cfg.Defaults.Model = "sonar"
	cfg.Defaults.Temperature = 0.7
	cfg.Defaults.MaxTokens = 4096
	cfg.Search.Mode = "academic"
	cfg.Search.ContextSize = "high"
	cfg.Search.Recency = "week"
	cfg.Output.Stream = true
	cfg.Output.ReturnImages = true
	cfg.Output.ReturnRelated = true

	opts := config.DefaultAnnotationOptions()
	opts.IncludeExamples = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := config.GenerateAnnotatedConfig(cfg, opts)
		if err != nil {
			b.Fatalf("GenerateAnnotatedConfig failed: %v", err)
		}
	}
}

// BenchmarkProfileOperations benchmarks profile management operations.
func BenchmarkProfileOperations(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create a config with profiles
	fixtureData, err := os.ReadFile(filepath.Join("testdata", "profile_config.yaml"))
	if err != nil {
		b.Fatalf("Failed to read fixture: %v", err)
	}
	if err := os.WriteFile(configPath, fixtureData, configFilePermission); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	loader := config.NewLoader()
	if err := loader.LoadFrom(configPath); err != nil {
		b.Fatalf("Failed to load config: %v", err)
	}

	pm := config.NewProfileManager(loader.Data())

	b.Run("ListProfiles", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = pm.ListProfiles()
		}
	})

	b.Run("LoadProfile", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := pm.LoadProfile("research")
			if err != nil {
				b.Fatalf("LoadProfile failed: %v", err)
			}
		}
	})

	b.Run("SetActiveProfile", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := pm.SetActiveProfile("creative"); err != nil {
				b.Fatalf("SetActiveProfile failed: %v", err)
			}
		}
	})
}

// BenchmarkMetadataRegistry benchmarks config metadata operations.
func BenchmarkMetadataRegistry(b *testing.B) {
	registry := config.NewMetadataRegistry()

	b.Run("GetAll", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			options := registry.GetAll()
			if len(options) < 29 {
				b.Fatalf("Expected at least 29 options, got %d", len(options))
			}
		}
	})

	b.Run("GetBySection", func(b *testing.B) {
		sections := []string{"defaults", "search", "output", "api"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, section := range sections {
				_ = registry.GetBySection(section)
			}
		}
	})
}

// BenchmarkFormatOptions benchmarks config options formatting.
func BenchmarkFormatOptions(b *testing.B) {
	registry := config.NewMetadataRegistry()
	options := registry.GetAll()

	formats := []string{"table", "json", "yaml"}

	for _, format := range formats {
		b.Run(format, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				output, err := config.FormatOptions(options, format)
				if err != nil {
					b.Fatalf("FormatOptions failed: %v", err)
				}
				if output == "" {
					b.Fatal("FormatOptions returned empty output")
				}
			}
		})
	}
}

// BenchmarkConcurrentConfigLoad benchmarks concurrent config loading.
func BenchmarkConcurrentConfigLoad(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	fixtureData, err := os.ReadFile(filepath.Join("testdata", "valid_config.yaml"))
	if err != nil {
		b.Fatalf("Failed to read fixture: %v", err)
	}
	if err := os.WriteFile(configPath, fixtureData, configFilePermission); err != nil {
		b.Fatalf("Failed to write config: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			loader := config.NewLoader()
			if err := loader.LoadFrom(configPath); err != nil {
				b.Fatalf("LoadFrom failed: %v", err)
			}
		}
	})
}

// BenchmarkConfigWithLargeProfiles benchmarks config with many profiles.
func BenchmarkConfigWithLargeProfiles(b *testing.B) {
	cfg := config.NewConfigData()
	cfg.Defaults.Model = "sonar"

	// Create 50 profiles
	for i := 0; i < 50; i++ {
		profileName := fmt.Sprintf("profile-%d", i)
		cfg.Profiles[profileName] = &config.Profile{
			Name:        profileName,
			Description: fmt.Sprintf("Test profile %d", i),
			Defaults: config.DefaultsConfig{
				Model:       "sonar",
				Temperature: float64(i) / 100.0,
				MaxTokens:   2000 + i*100,
			},
		}
	}

	b.Run("Validation", func(b *testing.B) {
		validator := config.NewValidator()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := validator.Validate(cfg); err != nil {
				b.Fatalf("Validate failed: %v", err)
			}
		}
	})

	b.Run("ProfileListing", func(b *testing.B) {
		pm := config.NewProfileManager(cfg)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			profiles := pm.ListProfiles()
			// ListProfiles always includes the default profile, so 50 created + 1 default = 51
			if len(profiles) != 51 {
				b.Fatalf("Expected 51 profiles (50 created + 1 default), got %d", len(profiles))
			}
		}
	})
}

// BenchmarkWizardOperations benchmarks wizard state operations.
func BenchmarkWizardOperations(b *testing.B) {
	b.Run("NewWizardState", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := NewWizardState()
			if w == nil {
				b.Fatal("NewWizardState returned nil")
			}
		}
	})

	b.Run("BuildConfiguration", func(b *testing.B) {
		w := NewWizardState()
		w.useCase = config.TemplateResearch
		w.selectedModel = "sonar"
		w.enableStream = true

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w.buildConfiguration()
			if w.config == nil {
				b.Fatal("buildConfiguration returned nil config")
			}
		}
	})
}

// BenchmarkMemoryAllocations tracks memory allocations for config operations.
func BenchmarkMemoryAllocations(b *testing.B) {
	b.Run("ConfigLoad", func(b *testing.B) {
		tempDir := b.TempDir()
		configPath := filepath.Join(tempDir, "config.yaml")

		fixtureData, err := os.ReadFile(filepath.Join("testdata", "valid_config.yaml"))
		if err != nil {
			b.Fatalf("Failed to read fixture: %v", err)
		}
		if err := os.WriteFile(configPath, fixtureData, configFilePermission); err != nil {
			b.Fatalf("Failed to write config: %v", err)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			loader := config.NewLoader()
			if err := loader.LoadFrom(configPath); err != nil {
				b.Fatalf("LoadFrom failed: %v", err)
			}
		}
	})

	b.Run("TemplateLoad", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cfg, err := config.LoadTemplate(config.TemplateResearch)
			if err != nil {
				b.Fatalf("LoadTemplate failed: %v", err)
			}
			if cfg == nil {
				b.Fatal("LoadTemplate returned nil")
			}
		}
	})

	b.Run("Validation", func(b *testing.B) {
		cfg := config.NewConfigData()
		cfg.Defaults.Model = "sonar"
		cfg.Defaults.Temperature = 0.7

		validator := config.NewValidator()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := validator.Validate(cfg); err != nil {
				b.Fatalf("Validate failed: %v", err)
			}
		}
	})
}
