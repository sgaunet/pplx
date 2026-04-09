package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"
)

// DiffEntry represents a single difference between two configurations.
type DiffEntry struct {
	Key   string `json:"key"`   // dot-notation key (e.g., "defaults.temperature")
	Left  string `json:"left"`  // formatted value from first config
	Right string `json:"right"` // formatted value from second config
}

// DiffConfigs compares two ConfigData structs and returns entries for keys whose
// values differ. Both configs are compared across all four sections (defaults,
// search, output, api). The returned slice is sorted alphabetically by key.
func DiffConfigs(a, b *ConfigData) []DiffEntry {
	keys := AllKeys()
	entries := make([]DiffEntry, 0, len(keys))

	for _, key := range keys {
		aVal, aErr := GetValue(a, key)
		bVal, bErr := GetValue(b, key)

		// Treat retrieval errors as empty/zero values represented by "".
		var aStr, bStr string
		if aErr == nil {
			aStr = fmt.Sprintf("%v", aVal)
		}
		if bErr == nil {
			bStr = fmt.Sprintf("%v", bVal)
		}

		if aStr != bStr {
			entries = append(entries, DiffEntry{Key: key, Left: aStr, Right: bStr})
		}
	}

	return entries
}

// FormatDiff renders diff entries in the requested format.
//
// Supported formats:
//   - "table" — aligned columns via text/tabwriter: Field, Left, Right
//   - "json"  — JSON array of [DiffEntry]
//
// Returns "No differences found." when entries is empty regardless of format.
func FormatDiff(entries []DiffEntry, format string) string {
	if len(entries) == 0 {
		return "No differences found."
	}

	switch format {
	case formatJSON:
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return fmt.Sprintf("json encoding error: %v", err)
		}
		return string(data)
	default: // "table" and any unrecognised format
		var buf bytes.Buffer
		w := tabwriter.NewWriter(&buf, 0, 0, defaultTabPadding, ' ', 0)
		_, _ = fmt.Fprintln(w, "Field\tLeft\tRight")
		_, _ = fmt.Fprintln(w, "-----\t----\t-----")
		for _, e := range entries {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", e.Key, e.Left, e.Right)
		}
		_ = w.Flush()
		return buf.String()
	}
}
