package main

import (
	"testing"
)

func TestExtractKey(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		column   int
		expected string
	}{
		{"first column", "apple\tbanana\tcherry", 1, "apple"},
		{"second column", "apple\tbanana\tcherry", 2, "banana"},
		{"third column", "apple\tbanana\tcherry", 3, "cherry"},
		{"column out of range", "apple\tbanana", 5, "apple\tbanana"},
		{"column zero", "apple\tbanana", 0, "apple\tbanana"},
		{"single column", "apple", 1, "apple"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			line := tt.line
			column := tt.column

			// When
			result := extractKey(line, column)

			// Then
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCompareByMonth(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"jan before feb", "January", "February", true},
		{"feb after jan", "February", "January", false},
		{"same month", "March", "March", false},
		{"dec before jan", "December", "January", false},
		{"case insensitive", "jun", "JUL", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			a := tt.a
			b := tt.b

			// When
			result := compareByMonth(a, b)

			// Then
			if result != tt.expected {
				t.Errorf("compareByMonth(%q, %q) = %v, expected %v", a, b, result, tt.expected)
			}
		})
	}
}

func TestCompareByNumber(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"1 less than 2", "1", "2", true},
		{"10 less than 100", "10", "100", true},
		{"100 not less than 10", "100", "10", false},
		{"equal numbers", "42", "42", false},
		{"negative numbers", "-5", "5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			a := tt.a
			b := tt.b

			// When
			result := compareByNumber(a, b)

			// Then
			if result != tt.expected {
				t.Errorf("compareByNumber(%q, %q) = %v, expected %v", a, b, result, tt.expected)
			}
		})
	}
}

func TestCompareByNumberWithSuffix(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"1k less than 2k", "1k", "2k", true},
		{"1k less than 1m", "1k", "1m", true},
		{"2m not less than 1k", "2m", "1k", false},
		{"1g greater than 1m", "1g", "1m", false},
		{"equal values", "1k", "1k", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			a := tt.a
			b := tt.b

			// When
			result := compareByNumberWithSuffix(a, b)

			// Then
			if result != tt.expected {
				t.Errorf("compareByNumberWithSuffix(%q, %q) = %v, expected %v", a, b, result, tt.expected)
			}
		})
	}
}

func TestCompareDefault(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"a before b", "apple", "banana", true},
		{"b after a", "banana", "apple", false},
		{"equal strings", "cherry", "cherry", false},
		{"case sensitive", "Apple", "apple", true},
		{"numbers as strings", "10", "2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			a := tt.a
			b := tt.b

			// When
			result := compareDefault(a, b)

			// Then
			if result != tt.expected {
				t.Errorf("compareDefault(%q, %q) = %v, expected %v", a, b, result, tt.expected)
			}
		})
	}
}

func TestParseSuffix(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		shouldError bool
	}{
		{"1k", "1k", 1024, false},
		{"2m", "2m", 2097152, false},
		{"3g", "3g", 3221225472, false},
		{"uppercase K", "5K", 5120, false},
		{"bad suffix", "10x", 0, true},
		{"no suffix", "100", 0, true},
		{"empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			input := tt.input

			// When
			result, err := parseSuffix(input)

			// Then
			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error for input %q, got none", input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", input, err)
				}
				if result != tt.expected {
					t.Errorf("parseSuffix(%q) = %d, expected %d", input, result, tt.expected)
				}
			}
		})
	}
}

func TestIsSorted(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected bool
	}{
		{"sorted strings", []string{"apple", "banana", "cherry"}, true},
		{"unsorted strings", []string{"banana", "apple", "cherry"}, false},
		{"single element", []string{"apple"}, true},
		{"empty slice", []string{}, true},
		{"reverse order", []string{"cherry", "banana", "apple"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			lines := tt.lines
			compare := compareDefault

			// When
			result := isSorted(lines, compare)

			// Then
			if result != tt.expected {
				t.Errorf("isSorted(%v) = %v, expected %v", lines, result, tt.expected)
			}
		})
	}
}
