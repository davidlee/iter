package config

import (
	"testing"

	"github.com/davidlee/vice/internal/zk"
)

func TestViceEnv_ZKIntegration(t *testing.T) {
	env, err := GetDefaultViceEnv()
	if err != nil {
		t.Fatalf("GetDefaultViceEnv() error = %v", err)
	}

	// ZK should be initialized
	if env.ZK == nil {
		t.Error("ViceEnv.ZK should be initialized")
	}

	// Should return zk tool name
	if env.ZK.Name() != "zk" {
		t.Errorf("ZK.Name() = %q, want %q", env.ZK.Name(), "zk")
	}

	// IsZKAvailable should work
	available := env.IsZKAvailable()
	// We don't test specific value since it depends on system PATH
	_ = available

	// WarnZKUnavailable should not panic
	env.WarnZKUnavailable()
}

func TestViceEnv_ZKList_Unavailable(t *testing.T) {
	env := &ViceEnv{
		ZK: &zk.ZKExecutable{}, // Create unavailable ZK instance
	}

	paths, err := env.ZKList("--tag", "vice:srs")
	if err == nil {
		t.Error("ZKList() should return error when ZK unavailable")
	}

	if paths != nil {
		t.Error("ZKList() should return nil paths when ZK unavailable")
	}

	expectedMsg := "zk not available - install from https://github.com/zk-org/zk"
	if err.Error() != expectedMsg {
		t.Errorf("ZKList() error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestViceEnv_ZKEdit_Unavailable(t *testing.T) {
	env := &ViceEnv{
		ZK: &zk.ZKExecutable{}, // Create unavailable ZK instance
	}

	err := env.ZKEdit("note.md")
	if err == nil {
		t.Error("ZKEdit() should return error when ZK unavailable")
	}

	expectedMsg := "zk not available - install from https://github.com/zk-org/zk"
	if err.Error() != expectedMsg {
		t.Errorf("ZKEdit() error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestViceEnv_GetZKNotebookDir(t *testing.T) {
	env, err := GetDefaultViceEnv()
	if err != nil {
		t.Fatalf("GetDefaultViceEnv() error = %v", err)
	}

	notebookDir := env.GetZKNotebookDir()
	flotsamDir := env.GetFlotsamDir()

	if notebookDir != flotsamDir {
		t.Errorf("GetZKNotebookDir() = %q, want %q", notebookDir, flotsamDir)
	}
}

func TestViceEnv_ValidateZKNotebook(t *testing.T) {
	env, err := GetDefaultViceEnv()
	if err != nil {
		t.Fatalf("GetDefaultViceEnv() error = %v", err)
	}

	// Should not error for non-existent notebook (NOOP validation)
	err = env.ValidateZKNotebook()
	if err != nil {
		t.Errorf("ValidateZKNotebook() should not error for non-existent notebook, got: %v", err)
	}
}

func TestViceEnv_IsZKAvailable_NilZK(t *testing.T) {
	env := &ViceEnv{
		ZK: nil,
	}

	if env.IsZKAvailable() {
		t.Error("IsZKAvailable() should return false when ZK is nil")
	}
}

func TestViceEnv_WarnZKUnavailable_NilZK(_ *testing.T) {
	env := &ViceEnv{
		ZK: nil,
	}

	// Should not panic with nil ZK
	env.WarnZKUnavailable()
}