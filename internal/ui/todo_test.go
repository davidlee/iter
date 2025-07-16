package ui

import (
	"testing"
	"time"

	"davidlee/vice/internal/config"
	"davidlee/vice/internal/models"
)

func TestTodoDashboard(t *testing.T) {
	// Create test ViceEnv
	env := &config.ViceEnv{
		ConfigDir:   "/tmp/test-config",
		Context:     "personal",
		ContextData: "testdata",
	}

	dashboard := NewTodoDashboard(env)

	// Test dashboard creation
	if dashboard == nil {
		t.Fatal("NewTodoDashboard returned nil")
	}

	if dashboard.env != env {
		t.Error("Dashboard env not set correctly")
	}
}

func TestGetStatusSymbol(t *testing.T) {
	td := &TodoDashboard{}

	tests := []struct {
		status   models.EntryStatus
		expected string
	}{
		{models.EntryCompleted, "✓"},
		{models.EntrySkipped, "⤫"},
		{models.EntryFailed, "✗"},
		{"pending", "○"},
		{"unknown", "?"},
	}

	for _, test := range tests {
		result := td.getStatusSymbol(test.status)
		if result != test.expected {
			t.Errorf("getStatusSymbol(%s) = %s, expected %s", test.status, result, test.expected)
		}
	}
}

func TestFormatValue(t *testing.T) {
	td := &TodoDashboard{}

	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, ""},
		{true, "true"},
		{false, "false"},
		{42, "42"},
		{"hello", "hello"},
		{time.Duration(5 * time.Minute), "5m0s"},
	}

	for _, test := range tests {
		result := td.formatValue(test.input)
		if result != test.expected {
			t.Errorf("formatValue(%v) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestTruncateString(t *testing.T) {
	td := &TodoDashboard{}

	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10", 10, "exactly10"},
		{"this is a very long string", 10, "this is..."},
		{"test", 3, "tes"},
		{"test", 2, "te"},
	}

	for _, test := range tests {
		result := td.truncateString(test.input, test.maxLen)
		if result != test.expected {
			t.Errorf("truncateString(%s, %d) = %s, expected %s", test.input, test.maxLen, result, test.expected)
		}
	}
}

func TestGetMarkdownCheckbox(t *testing.T) {
	td := &TodoDashboard{}

	tests := []struct {
		status   models.EntryStatus
		expected string
	}{
		{models.EntryCompleted, "- [x]"},
		{models.EntrySkipped, "- [-]"},
		{models.EntryFailed, "- [ ]"},
		{"pending", "- [ ]"},
		{"unknown", "- [ ]"},
	}

	for _, test := range tests {
		result := td.getMarkdownCheckbox(test.status)
		if result != test.expected {
			t.Errorf("getMarkdownCheckbox(%s) = %s, expected %s", test.status, result, test.expected)
		}
	}
}
