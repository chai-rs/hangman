package hangman

import "testing"

func TestIsAlphabet(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"single lowercase letter", "a", true},
		{"single uppercase letter", "A", true},
		{"multiple lowercase letters", "abc", true},
		{"multiple uppercase letters", "ABC", true},
		{"mixed case letters", "aBc", true},
		{"all lowercase alphabet", "abcdefghijklmnopqrstuvwxyz", true},
		{"numbers only", "123", false},
		{"letters with numbers", "a1", false},
		{"empty string", "", true},
		{"space only", " ", false},
		{"letters with space", "a b", false},
		{"special characters", "!@#", false},
		{"letters with special chars", "a!", false},
		{"underscore", "_", false},
		{"hyphen", "-", false},
		{"mixed alphanumeric", "abc123", false},
		{"unicode letter", "é", true},
		{"unicode letters", "café", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAlphabet(tt.input)
			if result != tt.expected {
				t.Errorf("IsAlphabet(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
