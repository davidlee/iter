package zk

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestZKTool_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		zkPath   string
		expected bool
	}{
		{
			name:     "zk available in PATH",
			zkPath:   "", // Will use NewZKTool() which searches PATH
			expected: zkExists(),
		},
		{
			name:     "zk not available",
			zkPath:   "/nonexistent/zk",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tool *ZKTool
			if tt.zkPath == "" {
				tool = NewZKTool()
			} else {
				tool = NewZKToolWithPath(tt.zkPath)
			}

			if got := tool.IsAvailable(); got != tt.expected {
				t.Errorf("ZKTool.IsAvailable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestZKTool_Name(t *testing.T) {
	tool := NewZKTool()
	if got := tool.Name(); got != "zk" {
		t.Errorf("ZKTool.Name() = %v, want %v", got, "zk")
	}
}

func TestZKTool_Version(t *testing.T) {
	if !zkExists() {
		t.Skip("zk not available, skipping version test")
	}

	tool := NewZKTool()
	version, err := tool.Version()
	if err != nil {
		t.Fatalf("ZKTool.Version() error = %v", err)
	}

	if version == "" {
		t.Error("ZKTool.Version() returned empty string")
	}

	// Version should contain some indication it's zk
	if !strings.Contains(strings.ToLower(version), "zk") {
		t.Errorf("ZKTool.Version() = %v, expected to contain 'zk'", version)
	}
}

func TestZKTool_Execute(t *testing.T) {
	if !zkExists() {
		t.Skip("zk not available, skipping execute test")
	}

	tool := NewZKTool()
	ctx := context.Background()

	// Test simple command that should succeed
	result, err := tool.Execute(ctx, "--help")
	if err != nil {
		t.Fatalf("ZKTool.Execute() error = %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("ZKTool.Execute() exit code = %v, want 0", result.ExitCode)
	}

	if len(result.Stdout) == 0 {
		t.Error("ZKTool.Execute() stdout is empty")
	}

	if result.Duration <= 0 {
		t.Error("ZKTool.Execute() duration should be positive")
	}
}

func TestZKTool_Execute_ToolNotFound(t *testing.T) {
	tool := NewZKToolWithPath("/nonexistent/zk")
	ctx := context.Background()

	_, err := tool.Execute(ctx, "--help")
	if err == nil {
		t.Fatal("ZKTool.Execute() expected error for nonexistent tool")
	}

	toolErr, ok := err.(*ToolError)
	if !ok {
		t.Fatalf("ZKTool.Execute() error type = %T, want *ToolError", err)
	}

	if toolErr.Type != ErrToolNotFound {
		t.Errorf("ToolError.Type = %v, want %v", toolErr.Type, ErrToolNotFound)
	}
}

func TestZKTool_Execute_WithTimeout(t *testing.T) {
	if !zkExists() {
		t.Skip("zk not available, skipping timeout test")
	}

	tool := NewZKTool()

	// Very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	_, err := tool.Execute(ctx, "--help")
	if err == nil {
		t.Error("ZKTool.Execute() expected timeout error")
	}
}

func TestZKCommand_List(t *testing.T) {
	tool := NewZKTool()
	cmd := tool.List()

	if cmd.command != "list" {
		t.Errorf("List() command = %v, want 'list'", cmd.command)
	}

	if len(cmd.args) != 1 || cmd.args[0] != "list" {
		t.Errorf("List() args = %v, want ['list']", cmd.args)
	}
}

func TestZKCommand_FluentInterface(t *testing.T) {
	tool := NewZKTool()

	cmd := tool.List().
		Format("json").
		Tag("vice:srs", "important").
		Match("concept").
		Limit(10).
		Interactive().
		NotebookDir("/path/to/notebook").
		NoInput()

	// Verify fluent interface returns same command
	if cmd.format != "json" {
		t.Errorf("Format() not set correctly, got %v", cmd.format)
	}

	expectedTags := []string{"vice:srs", "important"}
	if len(cmd.tags) != len(expectedTags) {
		t.Errorf("Tag() count = %v, want %v", len(cmd.tags), len(expectedTags))
	}

	for i, tag := range expectedTags {
		if cmd.tags[i] != tag {
			t.Errorf("Tag()[%d] = %v, want %v", i, cmd.tags[i], tag)
		}
	}

	if cmd.limit != 10 {
		t.Errorf("Limit() = %v, want 10", cmd.limit)
	}

	if !cmd.interactive {
		t.Error("Interactive() not set")
	}

	if cmd.notebookDir != "/path/to/notebook" {
		t.Errorf("NotebookDir() = %v, want '/path/to/notebook'", cmd.notebookDir)
	}

	if !cmd.noInput {
		t.Error("NoInput() not set")
	}
}

func TestZKCommand_New(t *testing.T) {
	tool := NewZKTool()

	cmd := tool.New().
		Title("My Note").
		Template("flotsam").
		Extra("tags=vice:srs,important").
		PrintPath().
		DryRun()

	// Check that arguments contain expected values
	expectedArgs := []string{
		"new", "--title", "My Note", "--template", "flotsam",
		"--extra", "tags=vice:srs,important", "--print-path", "--dry-run",
	}

	if len(cmd.args) != len(expectedArgs) {
		t.Errorf("New() args length = %v, want %v", len(cmd.args), len(expectedArgs))
	}

	for i, expected := range expectedArgs {
		if i < len(cmd.args) && cmd.args[i] != expected {
			t.Errorf("New() args[%d] = %v, want %v", i, cmd.args[i], expected)
		}
	}
}

func TestZKCommand_Edit(t *testing.T) {
	tool := NewZKTool()

	cmd := tool.Edit().
		Paths("note1.md", "note2.md").
		Tag("vice:srs").
		Interactive()

	if cmd.command != "edit" {
		t.Errorf("Edit() command = %v, want 'edit'", cmd.command)
	}

	expectedPaths := []string{"note1.md", "note2.md"}
	if len(cmd.paths) != len(expectedPaths) {
		t.Errorf("Paths() length = %v, want %v", len(cmd.paths), len(expectedPaths))
	}

	for i, path := range expectedPaths {
		if cmd.paths[i] != path {
			t.Errorf("Paths()[%d] = %v, want %v", i, cmd.paths[i], path)
		}
	}
}

func TestSanitizeArg(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean argument",
			input:    "normal-argument",
			expected: "normal-argument",
		},
		{
			name:     "argument with semicolon",
			input:    "arg;rm -rf /",
			expected: "argrm -rf /",
		},
		{
			name:     "argument with multiple dangerous chars",
			input:    "arg;$(rm -rf /)&",
			expected: "argrm -rf /",
		},
		{
			name:     "argument with pipes",
			input:    "arg | cat /etc/passwd",
			expected: "arg  cat /etc/passwd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeArg(tt.input); got != tt.expected {
				t.Errorf("sanitizeArg() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseZKPaths(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []string
	}{
		{
			name:     "empty input",
			input:    []byte(""),
			expected: []string{},
		},
		{
			name:     "single path",
			input:    []byte("/path/to/note.md"),
			expected: []string{"/path/to/note.md"},
		},
		{
			name:     "multiple paths",
			input:    []byte("/path/to/note1.md\n/path/to/note2.md\n/path/to/note3.md"),
			expected: []string{"/path/to/note1.md", "/path/to/note2.md", "/path/to/note3.md"},
		},
		{
			name:     "paths with trailing newline",
			input:    []byte("/path/to/note1.md\n/path/to/note2.md\n"),
			expected: []string{"/path/to/note1.md", "/path/to/note2.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseZKPaths(tt.input)
			if err != nil {
				t.Errorf("ParseZKPaths() error = %v", err)
				return
			}

			if len(got) != len(tt.expected) {
				t.Errorf("ParseZKPaths() length = %v, want %v", len(got), len(tt.expected))
				return
			}

			for i, path := range tt.expected {
				if got[i] != path {
					t.Errorf("ParseZKPaths()[%d] = %v, want %v", i, got[i], path)
				}
			}
		})
	}
}

func TestZKNote_HasTag(t *testing.T) {
	note := ZKNote{
		Tags: []string{"vice:srs", "vice:type:flashcard", "important"},
	}

	tests := []struct {
		tag      string
		expected bool
	}{
		{"vice:srs", true},
		{"vice:type:flashcard", true},
		{"important", true},
		{"nonexistent", false},
		{"vice:type:idea", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			if got := note.HasTag(tt.tag); got != tt.expected {
				t.Errorf("HasTag(%v) = %v, want %v", tt.tag, got, tt.expected)
			}
		})
	}
}

func TestZKNote_HasSRS(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected bool
	}{
		{
			name:     "has vice:srs tag",
			tags:     []string{"vice:srs", "important"},
			expected: true,
		},
		{
			name:     "no vice:srs tag",
			tags:     []string{"important", "concept"},
			expected: false,
		},
		{
			name:     "empty tags",
			tags:     []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := ZKNote{Tags: tt.tags}
			if got := note.HasSRS(); got != tt.expected {
				t.Errorf("HasSRS() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestZKNote_IsFlashcard(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected bool
	}{
		{
			name:     "is flashcard",
			tags:     []string{"vice:srs", "vice:type:flashcard"},
			expected: true,
		},
		{
			name:     "not flashcard",
			tags:     []string{"vice:srs", "vice:type:idea"},
			expected: false,
		},
		{
			name:     "no type tag",
			tags:     []string{"vice:srs"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := ZKNote{Tags: tt.tags}
			if got := note.IsFlashcard(); got != tt.expected {
				t.Errorf("IsFlashcard() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidateNotePath(t *testing.T) {
	// Create temporary directories for testing
	tmpDir := t.TempDir()
	notebookDir := tmpDir + "/notebook"
	err := os.MkdirAll(notebookDir, 0o750)
	require.NoError(t, err)

	tests := []struct {
		name        string
		notebookDir string
		notePath    string
		expectError bool
	}{
		{
			name:        "valid path inside notebook",
			notebookDir: notebookDir,
			notePath:    notebookDir + "/note.md",
			expectError: false,
		},
		{
			name:        "path outside notebook",
			notebookDir: notebookDir,
			notePath:    tmpDir + "/outside.md",
			expectError: true,
		},
		{
			name:        "path traversal attempt",
			notebookDir: notebookDir,
			notePath:    notebookDir + "/../outside.md",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotePath(tt.notebookDir, tt.notePath)
			if tt.expectError && err == nil {
				t.Error("ValidateNotePath() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateNotePath() unexpected error = %v", err)
			}
		})
	}
}

// Helper function to check if zk exists in PATH
func zkExists() bool {
	_, err := exec.LookPath("zk")
	return err == nil
}
