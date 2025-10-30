package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TestShortFlagAliases tests that all short flag aliases are properly defined
func TestShortFlagAliases(t *testing.T) {
	tests := []struct {
		name      string
		shortFlag string
		longFlag  string
		setupCmd  func(*cobra.Command)
	}{
		// Search flags
		{"search-domains", "d", "search-domains", addSearchFlags},
		{"search-recency", "r", "search-recency", addSearchFlags},

		// Response flags
		{"return-images", "i", "return-images", addResponseFlags},
		{"return-related", "q", "return-related", addResponseFlags},
		{"stream", "S", "stream", addResponseFlags},

		// Chat flags
		{"temperature", "t", "temperature", addChatFlags},
		{"max-tokens", "T", "max-tokens", addChatFlags},
		{"top-k", "k", "top-k", addChatFlags},

		// Format flags
		{"search-mode", "a", "search-mode", addFormatFlags},
		{"search-context-size", "c", "search-context-size", addFormatFlags},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			tt.setupCmd(cmd)

			// Check short flag exists
			shortFlag := cmd.PersistentFlags().ShorthandLookup(tt.shortFlag)
			if shortFlag == nil {
				t.Errorf("Short flag -%s not found", tt.shortFlag)
			}

			// Check long flag exists
			longFlag := cmd.PersistentFlags().Lookup(tt.longFlag)
			if longFlag == nil {
				t.Errorf("Long flag --%s not found", tt.longFlag)
			}

			// Verify they point to the same flag
			if shortFlag != nil && longFlag != nil {
				if shortFlag.Name != longFlag.Name {
					t.Errorf("Short flag -%s and long flag --%s do not point to the same flag",
						tt.shortFlag, tt.longFlag)
				}
			}
		})
	}
}

// TestBackwardCompatibility tests that long flags still work
func TestBackwardCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		longFlag string
		setupCmd func(*cobra.Command)
	}{
		{"search-domains-long", "search-domains", addSearchFlags},
		{"search-recency-long", "search-recency", addSearchFlags},
		{"return-images-long", "return-images", addResponseFlags},
		{"return-related-long", "return-related", addResponseFlags},
		{"stream-long", "stream", addResponseFlags},
		{"temperature-long", "temperature", addChatFlags},
		{"max-tokens-long", "max-tokens", addChatFlags},
		{"top-k-long", "top-k", addChatFlags},
		{"search-mode-long", "search-mode", addFormatFlags},
		{"search-context-size-long", "search-context-size", addFormatFlags},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			tt.setupCmd(cmd)

			flag := cmd.PersistentFlags().Lookup(tt.longFlag)
			if flag == nil {
				t.Errorf("Long flag --%s not found, backward compatibility broken", tt.longFlag)
			}
		})
	}
}

// TestExistingShortFlags tests that existing short flags are not affected
func TestExistingShortFlags(t *testing.T) {
	tests := []struct {
		name      string
		shortFlag string
		longFlag  string
		setupCmd  func(*cobra.Command)
	}{
		{"model", "m", "model", addChatFlags},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			tt.setupCmd(cmd)

			shortFlag := cmd.PersistentFlags().ShorthandLookup(tt.shortFlag)
			if shortFlag == nil {
				t.Errorf("Existing short flag -%s was removed", tt.shortFlag)
			}

			if shortFlag != nil && shortFlag.Name != tt.longFlag {
				t.Errorf("Short flag -%s does not map to expected long flag --%s",
					tt.shortFlag, tt.longFlag)
			}
		})
	}
}

// TestNoFlagConflicts tests that there are no conflicts between short flags
func TestNoFlagConflicts(t *testing.T) {
	cmd := &cobra.Command{}
	addChatFlags(cmd)
	addSearchFlags(cmd)
	addResponseFlags(cmd)
	addFormatFlags(cmd)
	addDateFlags(cmd)
	addImageFlags(cmd)
	addResearchFlags(cmd)

	// Collect all short flags
	shortFlags := make(map[string]string)
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Shorthand != "" {
			if existing, ok := shortFlags[flag.Shorthand]; ok {
				t.Errorf("Short flag conflict: -%s is used by both --%s and --%s",
					flag.Shorthand, existing, flag.Name)
			}
			shortFlags[flag.Shorthand] = flag.Name
		}
	})
}
