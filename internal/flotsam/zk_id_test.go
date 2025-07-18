package flotsam

import (
	"regexp"
	"strings"
	"testing"
)

func TestIDOptionsDefaults(t *testing.T) {
	opts := DefaultIDOptions()

	if opts.Length != 4 {
		t.Errorf("Expected default length 4, got %d", opts.Length)
	}

	if opts.Case != CaseLower {
		t.Errorf("Expected default case CaseLower, got %v", opts.Case)
	}

	expectedCharset := "0123456789abcdefghijklmnopqrstuvwxyz"
	if string(opts.Charset) != expectedCharset {
		t.Errorf("Expected default charset %s, got %s", expectedCharset, string(opts.Charset))
	}
}

func TestNewIDGeneratorLength(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"4-char", 4},
		{"8-char", 8},
		{"16-char", 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := IDOptions{
				Length:  tt.length,
				Charset: CharsetAlphanum,
				Case:    CaseLower,
			}

			generator := NewIDGenerator(opts)
			id := generator()

			if len(id) != tt.length {
				t.Errorf("Expected ID length %d, got %d", tt.length, len(id))
			}
		})
	}
}

func TestNewIDGeneratorCase(t *testing.T) {
	tests := []struct {
		name     string
		caseType Case
		pattern  string
		testChar rune
	}{
		{"lowercase", CaseLower, "^[0-9a-z]+$", 'A'},
		{"uppercase", CaseUpper, "^[0-9A-Z]+$", 'a'},
		{"mixed", CaseMixed, "^[0-9a-zA-Z]+$", 'A'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := IDOptions{
				Length:  100,          // Large length to increase chance of getting test character
				Charset: Charset("A"), // Single character to test case conversion
				Case:    tt.caseType,
			}

			generator := NewIDGenerator(opts)
			id := generator()

			matched, err := regexp.MatchString(tt.pattern, id)
			if err != nil {
				t.Fatalf("Regex error: %v", err)
			}

			if !matched {
				t.Errorf("ID %q doesn't match expected pattern %s", id, tt.pattern)
			}
		})
	}
}

func TestNewIDGeneratorCharset(t *testing.T) {
	tests := []struct {
		name    string
		charset Charset
		pattern string
	}{
		{"alphanum", CharsetAlphanum, "^[0-9a-z]+$"},
		{"hex", CharsetHex, "^[0-9a-f]+$"},
		{"letters", CharsetLetters, "^[a-z]+$"},
		{"numbers", CharsetNumbers, "^[0-9]+$"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := IDOptions{
				Length:  20,
				Charset: tt.charset,
				Case:    CaseLower,
			}

			generator := NewIDGenerator(opts)
			id := generator()

			matched, err := regexp.MatchString(tt.pattern, id)
			if err != nil {
				t.Fatalf("Regex error: %v", err)
			}

			if !matched {
				t.Errorf("ID %q doesn't match expected pattern %s for charset %v",
					id, tt.pattern, string(tt.charset))
			}
		})
	}
}

func TestNewIDGeneratorUniqueness(t *testing.T) {
	// AIDEV-NOTE: test uniqueness over many generations - critical for avoiding ID collisions
	// Note: With 4-char alphanumeric (36^4 = 1,679,616 possibilities), some collisions are expected
	opts := DefaultIDOptions()
	generator := NewIDGenerator(opts)

	seen := make(map[string]bool)
	iterations := 1000
	duplicates := 0

	for i := 0; i < iterations; i++ {
		id := generator()
		if seen[id] {
			duplicates++
			// Don't fail immediately - some collisions are statistically expected
		}
		seen[id] = true
	}

	// With 1000 iterations and 1.6M possibilities, expect very few duplicates (< 1%)
	maxExpectedDuplicates := 10 // Allow up to 1% collision rate
	if duplicates > maxExpectedDuplicates {
		t.Errorf("Too many duplicate IDs: %d (expected ≤ %d)", duplicates, maxExpectedDuplicates)
	}

	// Ensure we generated a reasonable number of unique IDs
	minExpectedUnique := iterations - maxExpectedDuplicates
	if len(seen) < minExpectedUnique {
		t.Errorf("Too few unique IDs: %d (expected ≥ %d)", len(seen), minExpectedUnique)
	}
}

func TestNewFlotsamIDGenerator(t *testing.T) {
	generator := NewFlotsamIDGenerator()
	id := generator()

	// Should match ZK's default format: 4-char alphanum lowercase
	if len(id) != 4 {
		t.Errorf("Expected Flotsam ID length 4, got %d", len(id))
	}

	matched, err := regexp.MatchString("^[0-9a-z]{4}$", id)
	if err != nil {
		t.Fatalf("Regex error: %v", err)
	}

	if !matched {
		t.Errorf("Flotsam ID %q doesn't match expected format", id)
	}
}

func TestFlotsamIDZKCompatibility(t *testing.T) {
	// AIDEV-NOTE: verify flotsam IDs are indistinguishable from ZK IDs
	generator := NewFlotsamIDGenerator()

	// Generate several IDs and verify they all match ZK format
	for i := 0; i < 10; i++ {
		id := generator()

		// ZK ID format: exactly 4 characters, alphanumeric, lowercase
		if len(id) != 4 {
			t.Errorf("ZK-incompatible ID length: expected 4, got %d", len(id))
		}

		for _, char := range id {
			if (char < '0' || char > '9') && (char < 'a' || char > 'z') {
				t.Errorf("ZK-incompatible character in ID %q: %c", id, char)
			}
		}

		// Verify no uppercase characters
		if strings.ToLower(id) != id {
			t.Errorf("ZK-incompatible uppercase in ID: %q", id)
		}
	}
}

func TestIDGeneratorPanicsOnInvalidLength(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for length 0")
		}
	}()

	opts := IDOptions{
		Length:  0,
		Charset: CharsetAlphanum,
		Case:    CaseLower,
	}

	NewIDGenerator(opts)
}

func TestIDGeneratorPanicsOnInvalidCase(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid case")
		}
	}()

	opts := IDOptions{
		Length:  4,
		Charset: CharsetAlphanum,
		Case:    Case(999), // Invalid case value
	}

	generator := NewIDGenerator(opts)
	generator() // Should panic when generating
}
