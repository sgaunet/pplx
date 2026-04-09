package config

import "strings"

// LevenshteinDistance computes the edit distance between two strings using
// the standard dynamic programming algorithm.
func LevenshteinDistance(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// prev holds costs for the previous row; curr for the current row.
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(
				prev[j]+1,      // deletion
				curr[j-1]+1,    // insertion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

// min3 returns the smallest of three integers.
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// SuggestEnum returns the element of valid that is closest to input
// (case-insensitive) and within maxDistance. Returns "" if no match qualifies.
func SuggestEnum(input string, valid []string, maxDistance int) string {
	lower := strings.ToLower(input)
	best, bestDist := "", maxDistance+1

	for _, v := range valid {
		d := LevenshteinDistance(lower, strings.ToLower(v))
		if d < bestDist {
			best, bestDist = v, d
		}
	}

	if bestDist <= maxDistance {
		return best
	}
	return ""
}
