package config

import "testing"

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"identical strings", "month", "month", 0},
		{"single char diff", "month", "monch", 1},
		{"monthly vs month", "monthly", "month", 2},
		{"completely different", "abc", "xyz", 3},
		{"both empty", "", "", 0},
		{"a empty", "", "abc", 3},
		{"b empty", "abc", "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LevenshteinDistance(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}

func TestSuggestEnum(t *testing.T) {
	recency := []string{"hour", "day", "week", "month", "year"}

	tests := []struct {
		name        string
		input       string
		valid       []string
		maxDistance int
		expected    string
	}{
		{"monthly suggests month", "monthly", recency, 2, "month"},
		{"no match xyz", "xyz", recency, 2, ""},
		{"wbe suggests web", "wbe", []string{"web", "academic"}, 2, "web"},
		{"exact match", "week", recency, 2, "week"},
		{"case insensitive", "MONTH", recency, 2, "month"},
		{"distance zero match", "day", recency, 0, "day"},
		{"too far away", "quarterly", recency, 2, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SuggestEnum(tt.input, tt.valid, tt.maxDistance)
			if got != tt.expected {
				t.Errorf("SuggestEnum(%q, ..., %d) = %q, want %q", tt.input, tt.maxDistance, got, tt.expected)
			}
		})
	}
}
