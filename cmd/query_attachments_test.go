package cmd

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestClassifyAttachment(t *testing.T) {
	tests := []struct {
		name      string
		entry     string
		wantImage bool
		wantURL   bool
		wantErr   bool
	}{
		{name: "local png", entry: "/tmp/photo.png", wantImage: true, wantURL: false, wantErr: false},
		{name: "local jpg uppercase ext", entry: "/tmp/photo.JPG", wantImage: true, wantURL: false, wantErr: false},
		{name: "local jpeg", entry: "./photo.jpeg", wantImage: true, wantURL: false, wantErr: false},
		{name: "local webp", entry: "photo.webp", wantImage: true, wantURL: false, wantErr: false},
		{name: "local gif", entry: "photo.gif", wantImage: true, wantURL: false, wantErr: false},
		{name: "local pdf", entry: "/tmp/doc.pdf", wantImage: false, wantURL: false, wantErr: false},
		{name: "local docx", entry: "/tmp/doc.docx", wantImage: false, wantURL: false, wantErr: false},
		{name: "local txt", entry: "notes.txt", wantImage: false, wantURL: false, wantErr: false},
		{name: "https image url", entry: "https://example.com/a.png", wantImage: true, wantURL: true, wantErr: false},
		{name: "https pdf url", entry: "https://arxiv.org/pdf/1706.03762.pdf", wantImage: false, wantURL: true, wantErr: false},
		{name: "http url rejected", entry: "http://example.com/a.png", wantErr: true},
		{name: "ftp url rejected", entry: "ftp://example.com/a.png", wantErr: true},
		{name: "empty", entry: "", wantErr: true},
		{name: "no extension", entry: "/tmp/something", wantErr: true},
		{name: "unknown extension", entry: "/tmp/foo.xyz", wantErr: true},
		{name: "url missing host", entry: "https:///foo.pdf", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotImage, gotURL, err := classifyAttachment(tt.entry)
			if (err != nil) != tt.wantErr {
				t.Fatalf("classifyAttachment(%q) err=%v, wantErr=%v", tt.entry, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotImage != tt.wantImage {
				t.Errorf("classifyAttachment(%q) isImage=%v, want %v", tt.entry, gotImage, tt.wantImage)
			}
			if gotURL != tt.wantURL {
				t.Errorf("classifyAttachment(%q) isURL=%v, want %v", tt.entry, gotURL, tt.wantURL)
			}
		})
	}
}

func TestValidateFiles(t *testing.T) {
	orig := globalOpts.Files
	defer func() { globalOpts.Files = orig }()

	// Create a real temp file for the positive path.
	tmp, err := os.CreateTemp(t.TempDir(), "pplx-test-*.png")
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	_ = tmp.Close()

	t.Run("empty list passes", func(t *testing.T) {
		globalOpts.Files = nil
		if err := validateFiles(); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	})

	t.Run("existing local file", func(t *testing.T) {
		globalOpts.Files = []string{tmp.Name()}
		if err := validateFiles(); err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	})

	t.Run("missing local file", func(t *testing.T) {
		globalOpts.Files = []string{filepath.Join(t.TempDir(), "nope.png")}
		if err := validateFiles(); err == nil {
			t.Fatal("expected validation error for missing file")
		}
	})

	t.Run("non-https URL", func(t *testing.T) {
		globalOpts.Files = []string{"http://example.com/x.pdf"}
		if err := validateFiles(); err == nil {
			t.Fatal("expected validation error for non-https URL")
		}
	})

	t.Run("unsupported extension", func(t *testing.T) {
		globalOpts.Files = []string{filepath.Join(t.TempDir(), "foo.xyz")}
		if err := validateFiles(); err == nil {
			t.Fatal("expected validation error for unsupported extension")
		}
	})
}

// TestBuildAllOptionsWithLocalImage exercises the full request-building path
// including perplexity-go's Validate() with a real local image attachment.
// This regression-guards the library fix that makes ValidateImageURL accept
// data:image/...;base64,... URIs produced by NewImageFileContent.
func TestBuildAllOptionsWithLocalImage(t *testing.T) {
	// Create a real, tiny PNG on disk so NewImageFileContent can encode it.
	pngPath := filepath.Join(t.TempDir(), "tiny.png")
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	if err := os.WriteFile(pngPath, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write png: %v", err)
	}

	// Snapshot and restore the globals we mutate. Other tests in this package
	// leak state into globalOpts, so we zero out every field that could trip
	// the library validator and restore them on exit.
	orig := *globalOpts
	t.Cleanup(func() { *globalOpts = orig })
	globalOpts.Files = []string{pngPath}
	globalOpts.UserPrompt = "describe this image"
	globalOpts.SearchRecency = ""
	globalOpts.ReturnImages = false
	globalOpts.ResponseFormatJSONSchema = ""
	globalOpts.ResponseFormatRegex = ""
	globalOpts.SearchDomains = nil
	globalOpts.ImageDomains = nil
	globalOpts.ImageFormats = nil
	globalOpts.SearchAfterDate = ""
	globalOpts.SearchBeforeDate = ""
	globalOpts.LastUpdatedAfter = ""
	globalOpts.LastUpdatedBefore = ""
	globalOpts.ReasoningEffort = ""

	req, err := buildAllOptions()
	if err != nil {
		t.Fatalf("buildAllOptions returned error with local image: %v", err)
	}
	if !req.IsMultimodal() {
		t.Fatal("expected multimodal request when --file is set")
	}
	if len(req.MultimodalMessages) == 0 {
		t.Fatal("expected MultimodalMessages to be populated")
	}
}
