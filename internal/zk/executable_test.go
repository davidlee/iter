package zk

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewZKExecutable(t *testing.T) {
	zk := NewZKExecutable()
	
	if zk == nil {
		t.Fatal("NewZKExecutable() returned nil")
	}
	
	if zk.Name() != "zk" {
		t.Errorf("Name() = %q, want %q", zk.Name(), "zk")
	}
	
	// Note: availability depends on system PATH, so we don't test specific value
	// but ensure the method doesn't panic
	_ = zk.Available()
	
	if zk.warned {
		t.Error("newly created ZKExecutable should not be warned yet")
	}
}

func TestZKExecutable_WarnIfUnavailable(t *testing.T) {
	// Create ZKExecutable with unavailable zk
	zk := &ZKExecutable{
		path:      "",
		available: false,
		warned:    false,
	}
	
	// First call should warn
	zk.WarnIfUnavailable()
	if !zk.warned {
		t.Error("WarnIfUnavailable() should set warned flag after first call")
	}
	
	// Second call should not warn again (tested by checking flag state)
	zk.WarnIfUnavailable()
	// No additional verification needed - the key is that warned flag prevents repeated warnings
}

func TestZKExecutable_Execute_Unavailable(t *testing.T) {
	zk := &ZKExecutable{
		path:      "",
		available: false,
		warned:    false,
	}
	
	result, err := zk.Execute("list")
	if err == nil {
		t.Error("Execute() should return error when zk is unavailable")
	}
	
	if result != nil {
		t.Error("Execute() should return nil result when zk is unavailable")
	}
	
	expectedMsg := "zk not available - install from https://github.com/zk-org/zk"
	if err.Error() != expectedMsg {
		t.Errorf("Execute() error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestZKExecutable_List_Unavailable(t *testing.T) {
	zk := &ZKExecutable{
		path:      "",
		available: false,
		warned:    false,
	}
	
	paths, err := zk.List("--tag", "vice:srs")
	if err == nil {
		t.Error("List() should return error when zk is unavailable")
	}
	
	if paths != nil {
		t.Error("List() should return nil paths when zk is unavailable")
	}
}

func TestZKExecutable_Edit_Unavailable(t *testing.T) {
	zk := &ZKExecutable{
		path:      "",
		available: false,
		warned:    false,
	}
	
	err := zk.Edit("note.md")
	if err == nil {
		t.Error("Edit() should return error when zk is unavailable")
	}
}

func TestZKExecutable_Edit_NoPaths(t *testing.T) {
	zk := &ZKExecutable{
		path:      "/usr/bin/zk", // Mock available zk
		available: true,
		warned:    false,
	}
	
	err := zk.Edit()
	if err == nil {
		t.Error("Edit() should return error when no paths specified")
	}
	
	expectedMsg := "no paths specified for editing"
	if err.Error() != expectedMsg {
		t.Errorf("Edit() error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestZKExecutable_GetLinkedNotes_Unavailable(t *testing.T) {
	zk := &ZKExecutable{
		path:      "",
		available: false,
		warned:    false,
	}
	
	backlinks, outbound, err := zk.GetLinkedNotes("note.md")
	if err == nil {
		t.Error("GetLinkedNotes() should return error when zk is unavailable")
	}
	
	if backlinks != nil || outbound != nil {
		t.Error("GetLinkedNotes() should return nil slices when zk is unavailable")
	}
}

func TestValidateZKConfig(t *testing.T) {
	// Test with non-existent config file
	err := ValidateZKConfig("/nonexistent/config.toml")
	if err != nil {
		t.Errorf("ValidateZKConfig() should not error for non-existent file, got: %v", err)
	}
	
	// Test with existing config file (create temporary file)
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	
	configContent := `[notebook]
dir = "."

[note]
filename = "{{id}}"
`
	
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	err = ValidateZKConfig(configPath)
	if err != nil {
		t.Errorf("ValidateZKConfig() should not error for existing valid file, got: %v", err)
	}
}

func TestFindZKNotebook(t *testing.T) {
	// Create temporary directory structure with .zk directory
	tmpDir := t.TempDir()
	notebookDir := filepath.Join(tmpDir, "notebook")
	zkDir := filepath.Join(notebookDir, ".zk")
	subDir := filepath.Join(notebookDir, "subdir")
	
	err := os.MkdirAll(zkDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create .zk directory: %v", err)
	}
	
	err = os.MkdirAll(subDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	
	// Test finding notebook from subdirectory
	foundDir, err := FindZKNotebook(subDir)
	if err != nil {
		t.Errorf("FindZKNotebook() error = %v, want nil", err)
	}
	
	if foundDir != notebookDir {
		t.Errorf("FindZKNotebook() = %q, want %q", foundDir, notebookDir)
	}
	
	// Test finding notebook from notebook root
	foundDir, err = FindZKNotebook(notebookDir)
	if err != nil {
		t.Errorf("FindZKNotebook() error = %v, want nil", err)
	}
	
	if foundDir != notebookDir {
		t.Errorf("FindZKNotebook() = %q, want %q", foundDir, notebookDir)
	}
	
	// Test not finding notebook
	nonNotebookDir := filepath.Join(tmpDir, "not-notebook")
	err = os.MkdirAll(nonNotebookDir, 0750)
	if err != nil {
		t.Fatalf("Failed to create non-notebook directory: %v", err)
	}
	
	_, err = FindZKNotebook(nonNotebookDir)
	if err == nil {
		t.Error("FindZKNotebook() should return error when .zk directory not found")
	}
}

func TestToolResult(t *testing.T) {
	result := &ToolResult{
		Stdout:   "test output",
		Stderr:   "test error",
		ExitCode: 0,
		Duration: 100 * time.Millisecond,
	}
	
	if result.Stdout != "test output" {
		t.Errorf("ToolResult.Stdout = %q, want %q", result.Stdout, "test output")
	}
	
	if result.Stderr != "test error" {
		t.Errorf("ToolResult.Stderr = %q, want %q", result.Stderr, "test error")
	}
	
	if result.ExitCode != 0 {
		t.Errorf("ToolResult.ExitCode = %d, want %d", result.ExitCode, 0)
	}
	
	if result.Duration != 100*time.Millisecond {
		t.Errorf("ToolResult.Duration = %v, want %v", result.Duration, 100*time.Millisecond)
	}
}