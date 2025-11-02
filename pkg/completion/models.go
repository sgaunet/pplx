// Package completion provides shell completion helpers and caching.
package completion

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheTTL is the time-to-live for cached model data.
const CacheTTL = 24 * time.Hour

// ModelCache represents cached model data.
type ModelCache struct {
	Models    []string  `json:"models"`
	UpdatedAt time.Time `json:"updated_at"`
}

// KnownModels returns the list of known Perplexity AI models.
// This list is based on https://docs.perplexity.ai/guides/model-cards
func KnownModels() []string {
	return []string{
		"sonar",                      // Default fast model
		"sonar-pro",                  // Pro model with enhanced capabilities
		"sonar-reasoning",            // Model with reasoning capabilities
		"sonar-deep-research",        // Deep research model with reasoning_effort
		"llama-3.1-sonar-small-128k-online",
		"llama-3.1-sonar-large-128k-online",
		"llama-3.1-sonar-huge-128k-online",
		"llama-3.1-8b-instruct",
		"llama-3.1-70b-instruct",
	}
}

// GetCacheDir returns the cache directory path.
func GetCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".cache", "pplx")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	return cacheDir, nil
}

// GetCachedModels retrieves models from cache if available and not expired.
func GetCachedModels() ([]string, error) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return nil, err
	}

	cachePath := filepath.Join(cacheDir, "models.json")

	// Check if cache file exists
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Cache doesn't exist, return known models
			return KnownModels(), nil
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Parse cache
	var cache ModelCache
	if err := json.Unmarshal(data, &cache); err != nil {
		// Invalid cache, return known models
		return KnownModels(), nil
	}

	// Check if cache is expired
	if time.Since(cache.UpdatedAt) > CacheTTL {
		// Cache expired, return known models
		return KnownModels(), nil
	}

	return cache.Models, nil
}

// SaveModelsToCache saves the model list to cache.
func SaveModelsToCache(models []string) error {
	cacheDir, err := GetCacheDir()
	if err != nil {
		return err
	}

	cachePath := filepath.Join(cacheDir, "models.json")

	cache := ModelCache{
		Models:    models,
		UpdatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// GetModels returns the list of available models, using cache when available.
func GetModels() []string {
	models, err := GetCachedModels()
	if err != nil {
		// Fallback to known models on error
		return KnownModels()
	}
	return models
}

// SearchModes returns valid search mode values.
func SearchModes() []string {
	return []string{"web", "academic"}
}

// RecencyValues returns valid recency filter values.
func RecencyValues() []string {
	return []string{"hour", "day", "week", "month", "year"}
}

// ContextSizes returns valid context size values.
func ContextSizes() []string {
	return []string{"low", "medium", "high"}
}

// ReasoningEfforts returns valid reasoning effort values for sonar-deep-research.
func ReasoningEfforts() []string {
	return []string{"low", "medium", "high"}
}

// ImageFormats returns common image format values.
func ImageFormats() []string {
	return []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg"}
}

// CommonDomains returns a list of common domains for suggestions.
func CommonDomains() []string {
	return []string{
		"github.com",
		"stackoverflow.com",
		"medium.com",
		"dev.to",
		"arxiv.org",
		"wikipedia.org",
		"reddit.com",
		"youtube.com",
	}
}
