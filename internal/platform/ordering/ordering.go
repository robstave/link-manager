package ordering

import (
	"fmt"
)

// Lexicographical string ordering utility
// Allows insertion between any two positions without re-indexing.
// Similar to Figma/Linear's fractional indexing using strings.

const (
	// Use lowercase letters a-z for simplicity and readability
	minChar = 'a'
	maxChar = 'z'
	midChar = 'n' // Middle character for initial positions
)

// Start returns the initial position string for the first item.
func Start() string {
	return string(midChar)
}

// End returns a position that sorts after all standard positions.
func End() string {
	return string(maxChar)
}

// Between generates a position string that sorts lexicographically between 'before' and 'after'.
// Empty strings are treated as boundaries (before="" means start, after="" means end).
func Between(before, after string) (string, error) {
	// Handle boundary cases
	if before == "" && after == "" {
		return Start(), nil
	}
	if before == "" {
		return beforeString(after), nil
	}
	if after == "" {
		return afterString(before), nil
	}

	// Validate: before must be < after
	if before >= after {
		return "", fmt.Errorf("invalid order: before (%q) must be less than after (%q)", before, after)
	}

	// Try to find a character between them at the first differing position
	minLen := len(before)
	if len(after) < minLen {
		minLen = len(after)
	}

	// Find first position where they differ
	pos := 0
	for pos < minLen && before[pos] == after[pos] {
		pos++
	}

	// Case 1: before is a prefix of after (e.g., "a" and "ab")
	if pos == len(before) {
		// Insert between last char of before and first char of suffix in after
		if pos < len(after) {
			// Try inserting a char between 'z' (implicit) and after[pos]
			if after[pos] > minChar {
				return before + string(after[pos]-1), nil
			}
			// after[pos] is 'a', need to extend: before + 'a' + midChar
			return before + string(minChar) + string(midChar), nil
		}
		// Shouldn't reach here if validation passed
		return "", fmt.Errorf("unexpected: before is prefix of after")
	}

	// Case 2: after is a prefix of before (shouldn't happen if before < after)
	if pos == len(after) {
		return "", fmt.Errorf("unexpected: after is prefix of before")
	}

	// Case 3: They differ at position 'pos'
	beforeChar := before[pos]
	afterChar := after[pos]

	// If there's a character between them, use it
	if afterChar-beforeChar > 1 {
		return before[:pos] + string(beforeChar+1), nil
	}

	// beforeChar and afterChar are adjacent (e.g., 'a' and 'b')
	// Need to look deeper or extend the string

	// If before has more characters after pos, we can increment within before's space
	if pos+1 < len(before) {
		// Try to increment the next position in before
		if before[pos+1] < maxChar {
			return before[:pos+1] + string(before[pos+1]+1), nil
		}
		// before[pos+1] is 'z', need to extend further
		return before + string(minChar), nil
	}

	// before ends at pos, after continues or also ends at pos
	// Insert: common_prefix + beforeChar + midChar
	return before[:pos] + string(beforeChar) + string(midChar), nil
}

// beforeString generates a position that comes before the given string.
func beforeString(s string) string {
	if s == "" || s[0] <= minChar {
		return string(minChar)
	}
	return string(s[0] - 1)
}

// afterString generates a position that comes after the given string.
func afterString(s string) string {
	// Append a middle character to ensure it comes after
	return s + string(midChar)
}

// NeedsRebalance checks if a position string is getting too long and should be rebalanced.
// This is a heuristic to detect when strings are growing inefficiently.
func NeedsRebalance(position string, threshold int) bool {
	if threshold <= 0 {
		threshold = 20 // Default threshold
	}
	return len(position) > threshold
}

// Rebalance generates new evenly-spaced position strings for a list of items.
// Returns a map of old position to new position.
func Rebalance(positions []string, spacing int) (map[string]string, error) {
	if spacing <= 0 {
		spacing = 10 // Default spacing between positions
	}

	result := make(map[string]string)
	for i, oldPos := range positions {
		// Generate positions like: "a", "k", "u", "be", "bo", ...
		newPos := generatePosition(i, spacing)
		result[oldPos] = newPos
	}

	return result, nil
}

// generatePosition creates a position string for index i with given spacing.
// Generates lexicographically increasing positions.
func generatePosition(index, spacing int) string {
	if index == 0 {
		return string(minChar)
	}

	// Use a simpler approach: generate position by repeatedly calling Between
	// This ensures proper lexicographic ordering
	pos := string(minChar)
	targetIndex := index * spacing

	for i := 1; i <= targetIndex; i++ {
		next, _ := Between(pos, "")
		pos = next
	}

	return pos
}
