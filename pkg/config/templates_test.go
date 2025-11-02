package config

import (
	"sync"
	"testing"
)

// TestLoadTemplate_ValidTemplates tests loading all valid templates.
func TestLoadTemplate_ValidTemplates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		templateName string
		wantErr      bool
	}{
		{
			name:         "load research template",
			templateName: TemplateResearch,
			wantErr:      false,
		},
		{
			name:         "load creative template",
			templateName: TemplateCreative,
			wantErr:      false,
		},
		{
			name:         "load news template",
			templateName: TemplateNews,
			wantErr:      false,
		},
		{
			name:         "load full-example template",
			templateName: TemplateFullExample,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config, err := LoadTemplate(tt.templateName)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && config == nil {
				t.Error("LoadTemplate() returned nil config")
			}
		})
	}
}

// TestLoadTemplate_InvalidTemplate tests loading an invalid template name.
func TestLoadTemplate_InvalidTemplate(t *testing.T) {
	t.Parallel()

	config, err := LoadTemplate("nonexistent")
	if err == nil {
		t.Error("LoadTemplate() expected error for invalid template, got nil")
	}

	if config != nil {
		t.Error("LoadTemplate() expected nil config for invalid template")
	}

	// Check that the error wraps ErrTemplateNotFound
	if !IsTemplateNotFoundError(err) {
		t.Errorf("LoadTemplate() expected error wrapping ErrTemplateNotFound, got %v", err)
	}
}

// TestLoadTemplate_ResearchConfiguration tests the research template has correct settings.
func TestLoadTemplate_ResearchConfiguration(t *testing.T) {
	t.Parallel()

	config, err := LoadTemplate(TemplateResearch)
	if err != nil {
		t.Fatalf("LoadTemplate() error = %v", err)
	}

	// Check temperature is low for factual responses
	if config.Defaults.Temperature >= 0.5 {
		t.Errorf("Research template temperature = %v, want < 0.5 for factual responses", config.Defaults.Temperature)
	}

	// Check academic mode if set
	if config.Search.Mode != "" && config.Search.Mode != "academic" {
		t.Errorf("Research template mode = %v, want 'academic'", config.Search.Mode)
	}

	// Check domains include scholarly sources
	if len(config.Search.Domains) == 0 {
		t.Error("Research template should have scholarly domains configured")
	}
}

// TestLoadTemplate_CreativeConfiguration tests the creative template has correct settings.
func TestLoadTemplate_CreativeConfiguration(t *testing.T) {
	t.Parallel()

	config, err := LoadTemplate(TemplateCreative)
	if err != nil {
		t.Fatalf("LoadTemplate() error = %v", err)
	}

	// Check temperature is high for creative responses
	if config.Defaults.Temperature <= 0.7 {
		t.Errorf("Creative template temperature = %v, want > 0.7 for creative responses", config.Defaults.Temperature)
	}

	// Check streaming is enabled
	if !config.Output.Stream {
		t.Error("Creative template should have streaming enabled")
	}
}

// TestLoadTemplate_NewsConfiguration tests the news template has correct settings.
func TestLoadTemplate_NewsConfiguration(t *testing.T) {
	t.Parallel()

	config, err := LoadTemplate(TemplateNews)
	if err != nil {
		t.Fatalf("LoadTemplate() error = %v", err)
	}

	// Check recency filter is set
	if config.Search.Recency == "" {
		t.Error("News template should have recency filter configured")
	}

	// Check news domains are configured
	if len(config.Search.Domains) == 0 {
		t.Error("News template should have news domains configured")
	}
}

// TestListTemplates tests that ListTemplates returns all expected templates.
func TestListTemplates(t *testing.T) {
	t.Parallel()

	templates := ListTemplates()

	// Check count
	expectedCount := 4
	if len(templates) != expectedCount {
		t.Errorf("ListTemplates() returned %d templates, want %d", len(templates), expectedCount)
	}

	// Check all expected templates are present
	expectedNames := []string{TemplateResearch, TemplateCreative, TemplateNews, TemplateFullExample}
	foundNames := make(map[string]bool)

	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("Template has empty name")
		}
		if tmpl.Description == "" {
			t.Errorf("Template %s has empty description", tmpl.Name)
		}
		if tmpl.UseCase == "" {
			t.Errorf("Template %s has empty use case", tmpl.Name)
		}
		foundNames[tmpl.Name] = true
	}

	for _, name := range expectedNames {
		if !foundNames[name] {
			t.Errorf("Expected template %s not found in ListTemplates()", name)
		}
	}
}

// TestGetTemplateDescription tests retrieving template descriptions.
func TestGetTemplateDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		templateName string
		wantErr      bool
	}{
		{
			name:         "get research description",
			templateName: TemplateResearch,
			wantErr:      false,
		},
		{
			name:         "get invalid template description",
			templateName: "nonexistent",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info, err := GetTemplateDescription(tt.templateName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplateDescription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if info == nil {
					t.Error("GetTemplateDescription() returned nil info")
				}
				if info.Name != tt.templateName {
					t.Errorf("GetTemplateDescription() name = %v, want %v", info.Name, tt.templateName)
				}
			}
		})
	}
}

// TestLoadTemplate_ConcurrentAccess tests concurrent template loading.
func TestLoadTemplate_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	const goroutines = 100
	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	templates := []string{TemplateResearch, TemplateCreative, TemplateNews, TemplateFullExample}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(templateName string) {
			defer wg.Done()
			_, err := LoadTemplate(templateName)
			if err != nil {
				errors <- err
			}
		}(templates[i%len(templates)])
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent LoadTemplate() error: %v", err)
	}
}

// TestLoadTemplate_Validation tests that loaded templates pass validation.
func TestLoadTemplate_Validation(t *testing.T) {
	t.Parallel()

	templates := []string{TemplateResearch, TemplateCreative, TemplateNews, TemplateFullExample}

	for _, name := range templates {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			config, err := LoadTemplate(name)
			if err != nil {
				t.Fatalf("LoadTemplate(%s) error = %v", name, err)
			}

			// Validate the configuration
			validator := NewValidator()
			if err := validator.Validate(config); err != nil {
				t.Errorf("Template %s failed validation: %v", name, err)
			}
		})
	}
}

// TestTemplateConstants tests that template constants are defined correctly.
func TestTemplateConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant string
		wantName string
	}{
		{
			name:     "research constant",
			constant: TemplateResearch,
			wantName: "research",
		},
		{
			name:     "creative constant",
			constant: TemplateCreative,
			wantName: "creative",
		},
		{
			name:     "news constant",
			constant: TemplateNews,
			wantName: "news",
		},
		{
			name:     "full-example constant",
			constant: TemplateFullExample,
			wantName: "full-example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.constant != tt.wantName {
				t.Errorf("Constant = %v, want %v", tt.constant, tt.wantName)
			}
		})
	}
}

// BenchmarkLoadTemplate benchmarks template loading performance.
func BenchmarkLoadTemplate(b *testing.B) {
	templates := []string{TemplateResearch, TemplateCreative, TemplateNews, TemplateFullExample}

	for _, name := range templates {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, err := LoadTemplate(name)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkListTemplates benchmarks ListTemplates performance.
func BenchmarkListTemplates(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ListTemplates()
	}
}
