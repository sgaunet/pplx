package completion

import "github.com/sgaunet/pplx/pkg/config"

// ConfigKeys returns all valid dot-notation configuration keys.
func ConfigKeys() []string {
	return config.AllKeys()
}

// ConfigValues returns valid values for enum-type configuration keys.
// For non-enum keys, returns nil (no completions).
func ConfigValues(key string) []string {
	switch key {
	case "search.recency":
		return RecencyValues()
	case "search.mode":
		return SearchModes()
	case "search.context_size":
		return ContextSizes()
	case "output.reasoning_effort":
		return ReasoningEfforts()
	case "output.stream", "output.return_images", "output.return_related", "output.json":
		return []string{"true", "false"}
	default:
		return nil
	}
}

// ProfileNames loads the config and returns all profile names.
func ProfileNames() []string {
	loader := config.NewLoader()
	if err := loader.Load(); err != nil {
		return []string{config.DefaultProfileName}
	}
	pm := config.NewProfileManager(loader.Data())
	return pm.ListProfiles()
}
