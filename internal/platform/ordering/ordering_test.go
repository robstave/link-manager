package ordering

import (
	"sort"
	"testing"
)

func TestStart(t *testing.T) {
	s := Start()
	if s != "n" {
		t.Errorf("Start() = %q, want %q", s, "n")
	}
}

func TestEnd(t *testing.T) {
	e := End()
	if e != "z" {
		t.Errorf("End() = %q, want %q", e, "z")
	}
}

func TestBetween_EmptyBoundaries(t *testing.T) {
	tests := []struct {
		name   string
		before string
		after  string
		want   string
	}{
		{"both empty", "", "", "n"},
		{"before empty", "", "p", "o"},
		{"after empty", "m", "", "mn"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Between(tt.before, tt.after)
			if err != nil {
				t.Fatalf("Between(%q, %q) error = %v", tt.before, tt.after, err)
			}
			if got != tt.want {
				t.Errorf("Between(%q, %q) = %q, want %q", tt.before, tt.after, got, tt.want)
			}
		})
	}
}

func TestBetween_SimpleInserts(t *testing.T) {
	tests := []struct {
		name   string
		before string
		after  string
	}{
		{"between a and c", "a", "c"},
		{"between a and b", "a", "b"},
		{"between m and o", "m", "o"},
		{"between n and z", "n", "z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Between(tt.before, tt.after)
			if err != nil {
				t.Fatalf("Between(%q, %q) error = %v", tt.before, tt.after, err)
			}
			// Verify got is between before and after
			if got <= tt.before || got >= tt.after {
				t.Errorf("Between(%q, %q) = %q, but %q should be between %q and %q",
					tt.before, tt.after, got, got, tt.before, tt.after)
			}
		})
	}
}

func TestBetween_PrefixCases(t *testing.T) {
	tests := []struct {
		name   string
		before string
		after  string
	}{
		{"before is prefix", "a", "ab"},
		{"before is prefix 2", "a", "az"},
		{"before is prefix 3", "abc", "abcd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Between(tt.before, tt.after)
			if err != nil {
				t.Fatalf("Between(%q, %q) error = %v", tt.before, tt.after, err)
			}
			if got <= tt.before || got >= tt.after {
				t.Errorf("Between(%q, %q) = %q, should be between", tt.before, tt.after, got)
			}
		})
	}
}

func TestBetween_InvalidOrder(t *testing.T) {
	tests := []struct {
		name   string
		before string
		after  string
	}{
		{"equal", "a", "a"},
		{"reversed", "b", "a"},
		{"reversed longer", "abc", "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Between(tt.before, tt.after)
			if err == nil {
				t.Errorf("Between(%q, %q) should return error for invalid order", tt.before, tt.after)
			}
		})
	}
}

func TestBetween_MultipleInsertions(t *testing.T) {
	// Start with two positions and insert multiple items between them
	positions := []string{"a", "z"}

	// Insert 10 items between a and z
	for i := 0; i < 10; i++ {
		before := positions[len(positions)-2]
		after := positions[len(positions)-1]

		mid, err := Between(before, after)
		if err != nil {
			t.Fatalf("Between(%q, %q) error = %v", before, after, err)
		}

		// Insert mid between before and after
		positions = append(positions[:len(positions)-1], mid, after)
	}

	// Verify all positions are in sorted order
	if !sort.StringsAreSorted(positions) {
		t.Errorf("positions not sorted: %v", positions)
	}

	t.Logf("Generated positions: %v", positions)
}

func TestBetween_SequentialInserts(t *testing.T) {
	// Build a sequence by repeatedly inserting after the last element
	var positions []string
	positions = append(positions, Start())

	for i := 0; i < 20; i++ {
		last := positions[len(positions)-1]
		next, err := Between(last, "")
		if err != nil {
			t.Fatalf("Between(%q, '') error = %v", last, err)
		}
		positions = append(positions, next)
	}

	// Verify sorted
	if !sort.StringsAreSorted(positions) {
		t.Errorf("positions not sorted: %v", positions)
	}

	t.Logf("Sequential positions: %v", positions)
}

func TestBetween_DenseInserts(t *testing.T) {
	// Insert between adjacent characters repeatedly
	before := "a"
	after := "b"

	var positions []string
	positions = append(positions, before)

	// Insert 5 items between a and b
	for i := 0; i < 5; i++ {
		mid, err := Between(before, after)
		if err != nil {
			t.Fatalf("Between(%q, %q) error = %v", before, after, err)
		}
		positions = append(positions, mid)
		before = mid
	}

	positions = append(positions, after)

	// Verify sorted
	if !sort.StringsAreSorted(positions) {
		t.Errorf("positions not sorted: %v", positions)
	}

	t.Logf("Dense positions between 'a' and 'b': %v", positions)
}

func TestNeedsRebalance(t *testing.T) {
	tests := []struct {
		name      string
		position  string
		threshold int
		want      bool
	}{
		{"short position", "a", 20, false},
		{"medium position", "abcdefghij", 20, false},
		{"long position", "abcdefghijklmnopqrstu", 20, true},
		{"custom threshold", "abcde", 3, true},
		{"default threshold", "abcdefghijklmnopqrstuv", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsRebalance(tt.position, tt.threshold)
			if got != tt.want {
				t.Errorf("NeedsRebalance(%q, %d) = %v, want %v",
					tt.position, tt.threshold, got, tt.want)
			}
		})
	}
}

func TestRebalance(t *testing.T) {
	// Create some positions that are unevenly distributed
	oldPositions := []string{"a", "aan", "aanm", "aanmz", "b"}

	newMap, err := Rebalance(oldPositions, 10)
	if err != nil {
		t.Fatalf("Rebalance error = %v", err)
	}

	if len(newMap) != len(oldPositions) {
		t.Errorf("Rebalance returned %d positions, want %d", len(newMap), len(oldPositions))
	}

	// Extract new positions and verify they're sorted
	var newPositions []string
	for _, oldPos := range oldPositions {
		newPos, ok := newMap[oldPos]
		if !ok {
			t.Errorf("missing new position for %q", oldPos)
			continue
		}
		newPositions = append(newPositions, newPos)
	}

	if !sort.StringsAreSorted(newPositions) {
		t.Errorf("rebalanced positions not sorted: %v", newPositions)
	}

	t.Logf("Original: %v", oldPositions)
	t.Logf("Rebalanced: %v", newPositions)
}

func TestGeneratePosition(t *testing.T) {
	// Test that generatePosition creates increasing sequences
	spacing := 10
	var positions []string

	for i := 0; i < 30; i++ {
		pos := generatePosition(i, spacing)
		positions = append(positions, pos)
	}

	if !sort.StringsAreSorted(positions) {
		t.Errorf("generated positions not sorted: %v", positions)
	}

	t.Logf("Generated positions: %v", positions)
}

func BenchmarkBetween(b *testing.B) {
	before := "a"
	after := "z"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Between(before, after)
	}
}

func BenchmarkBetweenDense(b *testing.B) {
	before := "a"
	after := "b"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mid, _ := Between(before, after)
		before = mid
	}
}
