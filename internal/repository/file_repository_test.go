package repository

import (
	"os"
	"path/filepath"
	"testing"

	"davidlee/vice/internal/config"
	"davidlee/vice/internal/models"
)

func createTestViceEnv(t *testing.T) *config.ViceEnv {
	tempDir := t.TempDir()
	
	env := &config.ViceEnv{
		ConfigDir:   filepath.Join(tempDir, "config"),
		DataDir:     filepath.Join(tempDir, "data"),
		StateDir:    filepath.Join(tempDir, "state"),
		CacheDir:    filepath.Join(tempDir, "cache"),
		Context:     "test",
		ContextData: filepath.Join(tempDir, "data", "test"),
		Contexts:    []string{"test", "work", "personal"},
	}
	
	if err := env.EnsureDirectories(); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}
	
	return env
}

func TestNewFileRepository(t *testing.T) {
	env := createTestViceEnv(t)
	
	repo := NewFileRepository(env)
	
	if repo.viceEnv != env {
		t.Error("ViceEnv not properly set")
	}
	if repo.dataLoaded {
		t.Error("Data should not be loaded initially")
	}
	if repo.habitParser == nil {
		t.Error("HabitParser should be initialized")
	}
	if repo.entryStorage == nil {
		t.Error("EntryStorage should be initialized")
	}
}

func TestGetCurrentContext(t *testing.T) {
	env := createTestViceEnv(t)
	repo := NewFileRepository(env)
	
	if repo.GetCurrentContext() != "test" {
		t.Errorf("GetCurrentContext() = %q, want %q", repo.GetCurrentContext(), "test")
	}
}

func TestListAvailableContexts(t *testing.T) {
	env := createTestViceEnv(t)
	repo := NewFileRepository(env)
	
	contexts := repo.ListAvailableContexts()
	expected := []string{"test", "work", "personal"}
	
	if len(contexts) != len(expected) {
		t.Errorf("ListAvailableContexts() length = %d, want %d", len(contexts), len(expected))
	}
	for i, ctx := range expected {
		if contexts[i] != ctx {
			t.Errorf("ListAvailableContexts()[%d] = %q, want %q", i, contexts[i], ctx)
		}
	}
}

func TestSwitchContext(t *testing.T) {
	env := createTestViceEnv(t)
	repo := NewFileRepository(env)
	
	// Test switching to valid context
	err := repo.SwitchContext("work")
	if err != nil {
		t.Fatalf("SwitchContext() failed: %v", err)
	}
	
	if repo.GetCurrentContext() != "work" {
		t.Errorf("Context not switched, got %q, want %q", repo.GetCurrentContext(), "work")
	}
	
	// Verify ContextData was updated
	expectedContextData := filepath.Join(env.DataDir, "work")
	if env.ContextData != expectedContextData {
		t.Errorf("ContextData not updated, got %q, want %q", env.ContextData, expectedContextData)
	}
	
	// Verify directory was created
	if _, err := os.Stat(env.ContextData); os.IsNotExist(err) {
		t.Error("Work context directory was not created")
	}
}

func TestSwitchContextInvalid(t *testing.T) {
	env := createTestViceEnv(t)
	repo := NewFileRepository(env)
	
	err := repo.SwitchContext("invalid")
	if err == nil {
		t.Error("SwitchContext() should fail for invalid context")
	}
	
	// Verify context didn't change
	if repo.GetCurrentContext() != "test" {
		t.Errorf("Context changed on invalid switch, got %q, want %q", repo.GetCurrentContext(), "test")
	}
}

func TestSwitchContextDataUnload(t *testing.T) {
	env := createTestViceEnv(t)
	repo := NewFileRepository(env)
	
	// Simulate some loaded data
	repo.dataLoaded = true
	repo.currentSchema = &models.Schema{Version: "1.0"}
	repo.currentEntries = &models.EntryLog{Version: "1.0"}
	
	// Switch context
	err := repo.SwitchContext("work")
	if err != nil {
		t.Fatalf("SwitchContext() failed: %v", err)
	}
	
	// Verify data was unloaded
	if repo.dataLoaded {
		t.Error("Data should be unloaded after context switch")
	}
	if repo.currentSchema != nil {
		t.Error("Current schema should be nil after context switch")
	}
	if repo.currentEntries != nil {
		t.Error("Current entries should be nil after context switch")
	}
}

func TestUnloadAllData(t *testing.T) {
	env := createTestViceEnv(t)
	repo := NewFileRepository(env)
	
	// Set up some loaded state
	repo.dataLoaded = true
	repo.currentSchema = &models.Schema{Version: "1.0"}
	repo.currentEntries = &models.EntryLog{Version: "1.0"}
	repo.currentChecklists = &models.ChecklistSchema{Version: "1.0"}
	repo.currentChecklistEntries = &models.ChecklistEntriesSchema{Version: "1.0"}
	
	// Unload data
	err := repo.UnloadAllData()
	if err != nil {
		t.Fatalf("UnloadAllData() failed: %v", err)
	}
	
	// Verify everything is cleared
	if repo.dataLoaded {
		t.Error("dataLoaded should be false after unload")
	}
	if repo.currentSchema != nil {
		t.Error("currentSchema should be nil after unload")
	}
	if repo.currentEntries != nil {
		t.Error("currentEntries should be nil after unload")
	}
	if repo.currentChecklists != nil {
		t.Error("currentChecklists should be nil after unload")
	}
	if repo.currentChecklistEntries != nil {
		t.Error("currentChecklistEntries should be nil after unload")
	}
}

func TestIsDataLoaded(t *testing.T) {
	env := createTestViceEnv(t)
	repo := NewFileRepository(env)
	
	// Initially no data loaded
	if repo.IsDataLoaded() {
		t.Error("IsDataLoaded() should return false initially")
	}
	
	// Set dataLoaded but no actual data
	repo.dataLoaded = true
	if repo.IsDataLoaded() {
		t.Error("IsDataLoaded() should return false when dataLoaded=true but no data")
	}
	
	// Add some data
	repo.currentSchema = &models.Schema{Version: "1.0"}
	if !repo.IsDataLoaded() {
		t.Error("IsDataLoaded() should return true when data is loaded")
	}
	
	// Unload data
	repo.UnloadAllData()
	if repo.IsDataLoaded() {
		t.Error("IsDataLoaded() should return false after unload")
	}
}

func TestLoadHabitsFileNotFound(t *testing.T) {
	env := createTestViceEnv(t)
	repo := NewFileRepository(env)
	
	// Try to load habits when file doesn't exist
	_, err := repo.LoadHabits()
	if err == nil {
		t.Error("LoadHabits() should fail when file doesn't exist")
	}
	
	// Verify it's a RepositoryError
	var repoErr *RepositoryError
	if !errorAs(err, &repoErr) {
		t.Errorf("Expected RepositoryError, got %T", err)
	}
	
	if repoErr.Operation != "LoadHabits" {
		t.Errorf("Expected operation 'LoadHabits', got %q", repoErr.Operation)
	}
	if repoErr.Context != "test" {
		t.Errorf("Expected context 'test', got %q", repoErr.Context)
	}
}

func TestRepositoryError(t *testing.T) {
	baseErr := &RepositoryError{
		Operation: "TestOp",
		Context:   "TestContext",
		Err:       os.ErrNotExist,
	}
	
	expected := "repository error in TestOp for context 'TestContext': file does not exist"
	if baseErr.Error() != expected {
		t.Errorf("Error() = %q, want %q", baseErr.Error(), expected)
	}
	
	if baseErr.Unwrap() != os.ErrNotExist {
		t.Errorf("Unwrap() = %v, want %v", baseErr.Unwrap(), os.ErrNotExist)
	}
}

// Helper function to check error type (similar to errors.As)
func errorAs(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	if repoErr, ok := err.(*RepositoryError); ok {
		if targetPtr, ok := target.(**RepositoryError); ok {
			*targetPtr = repoErr
			return true
		}
	}
	return false
}