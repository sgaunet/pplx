package config

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGenerateFieldComment(t *testing.T) {
	tests := []struct {
		name     string
		opt      *OptionMetadata
		indent   int
		contains []string
	}{
		{
			name: "basic comment with all fields",
			opt: &OptionMetadata{
				Name:            "model",
				Type:            "string",
				Description:     "AI model to use for queries",
				Default:         "sonar",
				ValidationRules: []string{"sonar", "sonar-pro", "gpt-4"},
				Example:         "sonar-pro",
				EnvVar:          "PPLX_MODEL",
				Required:        false,
			},
			indent: 0,
			contains: []string{
				"AI model to use for queries",
				"Type: string",
				"Default: sonar",
				"Valid values:",
				"Example: sonar-pro",
				"Env var: PPLX_MODEL",
			},
		},
		{
			name: "required field",
			opt: &OptionMetadata{
				Name:        "api_key",
				Type:        "string",
				Description: "API key for authentication",
				Required:    true,
			},
			indent: 0,
			contains: []string{
				"API key for authentication",
				"Required: true",
			},
		},
		{
			name: "with indentation",
			opt: &OptionMetadata{
				Name:        "temperature",
				Description: "Controls randomness in responses",
			},
			indent: 2,
			contains: []string{
				"  # Controls randomness in responses",
			},
		},
		{
			name:     "nil option",
			opt:      nil,
			indent:   0,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFieldComment(tt.opt, tt.indent)

			if tt.opt == nil {
				if result != "" {
					t.Errorf("expected empty string for nil option, got %q", result)
				}
				return
			}

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected comment to contain %q, got:\n%s", expected, result)
				}
			}

			// Verify all lines start with comment prefix (with proper indentation)
			lines := strings.Split(result, "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				expectedPrefix := strings.Repeat(" ", tt.indent) + commentPrefix
				if !strings.HasPrefix(line, expectedPrefix) {
					t.Errorf("line %q does not start with expected prefix %q", line, expectedPrefix)
				}
			}
		})
	}
}

func TestWrapComment(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		indent      int
		wantLines   int
		checkPrefix bool
	}{
		{
			name:        "short text no wrap",
			text:        "This is a short comment",
			indent:      0,
			wantLines:   1,
			checkPrefix: true,
		},
		{
			name:        "long text wraps",
			text:        "This is a very long comment that should definitely wrap to multiple lines because it exceeds the maximum comment width that we have configured",
			indent:      0,
			wantLines:   2, // Should wrap to at least 2 lines
			checkPrefix: true,
		},
		{
			name:        "with indentation",
			text:        "Comment with indentation",
			indent:      4,
			wantLines:   1,
			checkPrefix: true,
		},
		{
			name:        "empty text",
			text:        "",
			indent:      0,
			wantLines:   0,
			checkPrefix: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := wrapComment(tt.text, tt.indent)

			if len(lines) < tt.wantLines {
				t.Errorf("wrapComment() returned %d lines, want at least %d", len(lines), tt.wantLines)
			}

			if tt.checkPrefix {
				expectedPrefix := strings.Repeat(" ", tt.indent) + commentPrefix
				for _, line := range lines {
					if !strings.HasPrefix(line, expectedPrefix) {
						t.Errorf("line %q does not start with expected prefix %q", line, expectedPrefix)
					}
				}
			}

			// Verify no line exceeds max width
			for _, line := range lines {
				if len(line) > maxCommentWidth+tt.indent+len(commentPrefix) {
					t.Errorf("line exceeds max width: %q (length: %d)", line, len(line))
				}
			}
		})
	}
}

func TestCommentWrapping(t *testing.T) {
	// Test that very long words don't cause infinite loops
	longWord := strings.Repeat("a", 100)
	lines := wrapComment(longWord, 0)

	if len(lines) == 0 {
		t.Error("wrapComment() should handle very long words")
	}

	// At least one line should contain the long word
	found := false
	for _, line := range lines {
		if strings.Contains(line, longWord) {
			found = true
			break
		}
	}
	if !found {
		t.Error("long word was lost during wrapping")
	}
}

func BenchmarkGenerateFieldComment(b *testing.B) {
	opt := &OptionMetadata{
		Name:            "model",
		Type:            "string",
		Description:     "AI model to use for queries with detailed explanation of usage",
		Default:         "sonar",
		ValidationRules: []string{"sonar", "sonar-pro", "gpt-4", "claude-3"},
		Example:         "sonar-pro",
		EnvVar:          "PPLX_MODEL",
		Required:        false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateFieldComment(opt, 0)
	}
}

func BenchmarkWrapComment(b *testing.B) {
	text := "This is a long comment that needs to be wrapped to multiple lines because it exceeds the maximum comment width"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapComment(text, 0)
	}
}

func TestGenerateSectionHeader(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		style        HeaderStyle
		indent       int
		checkContent []string
	}{
		{
			name:         "box style",
			title:        "Section Title",
			style:        HeaderStyleBox,
			indent:       0,
			checkContent: []string{"╭", "╮", "│", "╰", "Section Title"},
		},
		{
			name:         "line style",
			title:        "Section Title",
			style:        HeaderStyleLine,
			indent:       0,
			checkContent: []string{"═", "Section Title"},
		},
		{
			name:         "minimal style",
			title:        "Section Title",
			style:        HeaderStyleMinimal,
			indent:       0,
			checkContent: []string{"Section Title", "-"},
		},
		{
			name:         "box style with indent",
			title:        "Nested Section",
			style:        HeaderStyleBox,
			indent:       2,
			checkContent: []string{"  # ╭", "Nested Section"},
		},
		{
			name:         "default to box style",
			title:        "Default",
			style:        HeaderStyle("invalid"),
			indent:       0,
			checkContent: []string{"╭", "Default"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSectionHeader(tt.title, tt.style, tt.indent)

			for _, expected := range tt.checkContent {
				if !strings.Contains(result, expected) {
					t.Errorf("expected header to contain %q, got:\n%s", expected, result)
				}
			}

			// Verify result has multiple lines
			lines := strings.Split(result, "\n")
			if len(lines) < 2 {
				t.Errorf("expected at least 2 lines, got %d", len(lines))
			}

			// Verify all lines start with comment prefix (with proper indentation)
			expectedPrefix := strings.Repeat(" ", tt.indent) + commentPrefix
			for i, line := range lines {
				if line == "" && i == len(lines)-1 {
					continue // Skip empty last line from trailing newline
				}
				if !strings.HasPrefix(line, expectedPrefix) {
					t.Errorf("line %d %q does not start with expected prefix %q", i, line, expectedPrefix)
				}
			}
		})
	}
}

func TestGenerateBoxHeader(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		indent    string
		minLength int
	}{
		{
			name:      "short title",
			title:     "Test",
			indent:    "",
			minLength: minHeaderWidth,
		},
		{
			name:      "long title",
			title:     "This is a very long section title that exceeds minimum width",
			indent:    "",
			minLength: len("This is a very long section title that exceeds minimum width") + 4,
		},
		{
			name:      "with indentation",
			title:     "Indented",
			indent:    "  ",
			minLength: minHeaderWidth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateBoxHeader(tt.title, tt.indent)
			lines := strings.Split(result, "\n")

			// Should have exactly 3 lines
			if len(lines) != 3 {
				t.Errorf("expected 3 lines, got %d", len(lines))
			}

			// Check for box characters
			if !strings.Contains(lines[0], "╭") || !strings.Contains(lines[0], "╮") {
				t.Errorf("top line missing box corners: %s", lines[0])
			}

			if !strings.Contains(lines[1], "│") {
				t.Errorf("middle line missing box sides: %s", lines[1])
			}

			if !strings.Contains(lines[2], "╰") || !strings.Contains(lines[2], "╯") {
				t.Errorf("bottom line missing box corners: %s", lines[2])
			}

			// Check title is present
			if !strings.Contains(lines[1], tt.title) {
				t.Errorf("title %q not found in header", tt.title)
			}
		})
	}
}

func TestGenerateLineHeader(t *testing.T) {
	title := "Test Section"
	result := generateLineHeader(title, "")
	lines := strings.Split(result, "\n")

	// Should have exactly 3 lines
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// Check for line separators
	if !strings.Contains(lines[0], "═") {
		t.Errorf("expected top separator to contain ═")
	}

	if !strings.Contains(lines[2], "═") {
		t.Errorf("expected bottom separator to contain ═")
	}

	// Check title is present
	if !strings.Contains(lines[1], title) {
		t.Errorf("title %q not found in header", title)
	}
}

func TestGenerateMinimalHeader(t *testing.T) {
	title := "Test Section"
	result := generateMinimalHeader(title, "")
	lines := strings.Split(result, "\n")

	// Should have exactly 2 lines
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Check title is present
	if !strings.Contains(lines[0], title) {
		t.Errorf("title %q not found in header", title)
	}

	// Check underline
	if !strings.Contains(lines[1], "-") {
		t.Errorf("expected underline to contain -")
	}
}

func BenchmarkGenerateSectionHeader(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateSectionHeader("Configuration Section", HeaderStyleBox, 0)
	}
}

func TestGenerateAnnotatedConfig(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *ConfigData
		opts         AnnotationOptions
		checkContent []string
		checkParse   bool
	}{
		{
			name:         "nil config uses defaults",
			cfg:          nil,
			opts:         DefaultAnnotationOptions(),
			checkContent: []string{"Perplexity CLI Configuration", "defaults:", "search:", "output:", "api:"},
			checkParse:   true,
		},
		{
			name: "custom config with values",
			cfg: &ConfigData{
				Defaults: DefaultsConfig{
					Model:       "sonar-pro",
					Temperature: 0.7,
					MaxTokens:   2000,
				},
				Search: SearchConfig{
					Mode:    "academic",
					Recency: "week",
				},
				Output: OutputConfig{
					Stream: true,
					JSON:   true,
				},
				API: APIConfig{
					Key:     "test-key",
					BaseURL: "https://api.test.com",
				},
			},
			opts: DefaultAnnotationOptions(),
			checkContent: []string{
				"sonar-pro",
				"0.7",
				"2000",
				"academic",
				"week",
				"stream: true",
				"json: true",
				"test-key",
				"https://api.test.com",
			},
			checkParse: true,
		},
		{
			name: "with example profiles",
			cfg:  NewConfigData(),
			opts: AnnotationOptions{
				IncludeExamples:     true,
				HeaderStyle:         HeaderStyleBox,
				IncludeDescriptions: true,
			},
			checkContent: []string{
				"Example Profiles",
				"Example: research",
				"Example: creative",
				"Example: news",
			},
			checkParse: true,
		},
		{
			name: "without descriptions",
			cfg:  NewConfigData(),
			opts: AnnotationOptions{
				IncludeExamples:     false,
				HeaderStyle:         HeaderStyleLine,
				IncludeDescriptions: false,
			},
			checkContent: []string{
				"defaults:",
				"search:",
				"output:",
				"api:",
			},
			checkParse: true,
		},
		{
			name: "with profiles",
			cfg: &ConfigData{
				Defaults: DefaultsConfig{
					Model: "sonar",
				},
				Profiles: map[string]*Profile{
					"test-profile": {
						Defaults: DefaultsConfig{
							Model:       "sonar-pro",
							Temperature: 0.5,
						},
					},
				},
				ActiveProfile: "test-profile",
			},
			opts: DefaultAnnotationOptions(),
			checkContent: []string{
				"Profiles",
				"test-profile:",
				"active_profile: test-profile",
			},
			checkParse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateAnnotatedConfig(tt.cfg, tt.opts)
			if err != nil {
				t.Fatalf("GenerateAnnotatedConfig() error = %v", err)
			}

			// Check for expected content
			for _, expected := range tt.checkContent {
				if !strings.Contains(result, expected) {
					t.Errorf("expected output to contain %q", expected)
				}
			}

			// Test YAML parsing if requested
			if tt.checkParse {
				var parsed ConfigData
				// Remove comment lines for parsing
				yamlLines := []string{}
				for _, line := range strings.Split(result, "\n") {
					if !strings.HasPrefix(strings.TrimSpace(line), "#") && line != "" {
						yamlLines = append(yamlLines, line)
					}
				}
				yamlContent := strings.Join(yamlLines, "\n")

				err := yaml.Unmarshal([]byte(yamlContent), &parsed)
				if err != nil {
					t.Logf("YAML content that failed to parse:\n%s", yamlContent)
					t.Errorf("failed to parse generated YAML: %v", err)
				}
			}
		})
	}
}

func TestGenerateAnnotatedConfig_Validation(t *testing.T) {
	// Generate an annotated config
	cfg := NewConfigData()
	cfg.Defaults.Model = "sonar"
	cfg.Defaults.Temperature = 0.5

	result, err := GenerateAnnotatedConfig(cfg, DefaultAnnotationOptions())
	if err != nil {
		t.Fatalf("GenerateAnnotatedConfig() error = %v", err)
	}

	// Remove comment lines
	yamlLines := []string{}
	for _, line := range strings.Split(result, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") && line != "" {
			yamlLines = append(yamlLines, line)
		}
	}
	yamlContent := strings.Join(yamlLines, "\n")

	// Parse the YAML
	var parsed ConfigData
	if err := yaml.Unmarshal([]byte(yamlContent), &parsed); err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}

	// Validate the parsed config
	validator := NewValidator()
	if err := validator.Validate(&parsed); err != nil {
		t.Errorf("validation failed on generated config: %v", err)
	}

	// Verify values match
	if parsed.Defaults.Model != cfg.Defaults.Model {
		t.Errorf("model mismatch: got %q, want %q", parsed.Defaults.Model, cfg.Defaults.Model)
	}
	if parsed.Defaults.Temperature != cfg.Defaults.Temperature {
		t.Errorf("temperature mismatch: got %v, want %v", parsed.Defaults.Temperature, cfg.Defaults.Temperature)
	}
}

func TestGenerateAnnotatedConfig_AllSections(t *testing.T) {
	cfg := NewConfigData()
	result, err := GenerateAnnotatedConfig(cfg, DefaultAnnotationOptions())
	if err != nil {
		t.Fatalf("GenerateAnnotatedConfig() error = %v", err)
	}

	// Verify all 4 main sections are present
	requiredSections := []string{
		"# ╭",                    // Box header start
		"Defaults",               // Section title
		"defaults:",              // YAML key
		"Search Options",         // Section title
		"search:",                // YAML key
		"Output Options",         // Section title
		"output:",                // YAML key
		"API Configuration",      // Section title
		"api:",                   // YAML key
		"model:",                 // A defaults field
		"temperature:",           // Another defaults field
		"mode:",                  // A search field
		"stream:",                // An output field
		"key:",                   // An api field
		"Type:",                  // Field documentation
		"Default:",               // Default value documentation
	}

	for _, section := range requiredSections {
		if !strings.Contains(result, section) {
			t.Errorf("expected output to contain section/field %q", section)
		}
	}
}

func TestDefaultAnnotationOptions(t *testing.T) {
	opts := DefaultAnnotationOptions()

	if opts.IncludeExamples {
		t.Error("expected IncludeExamples to be false by default")
	}
	if opts.HeaderStyle != HeaderStyleBox {
		t.Errorf("expected HeaderStyle to be Box, got %v", opts.HeaderStyle)
	}
	if !opts.IncludeDescriptions {
		t.Error("expected IncludeDescriptions to be true by default")
	}
}

func BenchmarkGenerateAnnotatedConfig(b *testing.B) {
	cfg := NewConfigData()
	cfg.Defaults.Model = "sonar"
	cfg.Defaults.Temperature = 0.7
	opts := DefaultAnnotationOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateAnnotatedConfig(cfg, opts)
	}
}
