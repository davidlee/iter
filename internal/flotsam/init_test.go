package flotsam

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/davidlee/vice/internal/config"
	"github.com/davidlee/vice/internal/zk"
)

func TestEnsureFlotsamEnvironment(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "flotsam-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name          string
		zkAvailable   bool
		flotsamExists bool
		expectInit    bool
		expectError   bool
	}{
		{
			name:          "AutoInit_ZKAvailable_NoFlotsamDir",
			zkAvailable:   true,
			flotsamExists: false,
			expectInit:    true,
			expectError:   false,
		},
		{
			name:          "NoInit_ZKUnavailable_NoFlotsamDir",
			zkAvailable:   false,
			flotsamExists: false,
			expectInit:    false,
			expectError:   false,
		},
		{
			name:          "NoInit_FlotsamExists",
			zkAvailable:   true,
			flotsamExists: true,
			expectInit:    false,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			testContextDir := filepath.Join(tempDir, tt.name)
			if err := os.MkdirAll(testContextDir, 0o750); err != nil {
				t.Fatalf("Failed to create test context dir: %v", err)
			}

			// Pre-create flotsam directory if test requires it
			if tt.flotsamExists {
				flotsamDir := filepath.Join(testContextDir, "flotsam")
				if err := os.MkdirAll(flotsamDir, 0o750); err != nil {
					t.Fatalf("Failed to create existing flotsam dir: %v", err)
				}
			}

			// Create mock ZK tool
			var mockZK zk.ZKTool
			if tt.zkAvailable {
				mockZK = zk.NewMockZKTool()
			} else {
				mockZK = zk.NewUnavailableMockZKTool()
			}

			// Create a test environment
			env := &config.ViceEnv{
				ContextData: testContextDir,
				Context:     "test",
			}

			// Run the testable function with injected mock
			err := EnsureFlotsamEnvironmentWithZK(env, mockZK)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check initialization expectation
			flotsamDir := filepath.Join(testContextDir, "flotsam")
			flotsamExists := dirExists(flotsamDir)

			if tt.expectInit && !flotsamExists {
				t.Error("Expected flotsam directory to be created but it wasn't")
			}
			if !tt.expectInit && tt.flotsamExists && !flotsamExists {
				t.Error("Expected existing flotsam directory to remain but it's missing")
			}

			// If ZK was available and init expected, verify ZK init was called
			if tt.expectInit && tt.zkAvailable {
				if mockTool, ok := mockZK.(*zk.MockZKTool); ok {
					if !mockTool.InitCalled {
						t.Error("Expected ZK init to be called but it wasn't")
					}
				}
			}
		})
	}
}

func TestIsFlotsamInitialized(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "flotsam-init-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	tests := []struct {
		name           string
		setupDirs      []string
		expectedResult bool
	}{
		{
			name:           "Complete_Setup",
			setupDirs:      []string{"flotsam", "flotsam/.zk"},
			expectedResult: true,
		},
		{
			name:           "Missing_ZK_Dir",
			setupDirs:      []string{"flotsam"},
			expectedResult: false,
		},
		{
			name:           "Missing_Flotsam_Dir",
			setupDirs:      []string{},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test context directory
			testContextDir := filepath.Join(tempDir, tt.name)
			if err := os.MkdirAll(testContextDir, 0o750); err != nil {
				t.Fatalf("Failed to create test context dir: %v", err)
			}

			// Create required directories
			for _, dirPath := range tt.setupDirs {
				fullPath := filepath.Join(testContextDir, dirPath)
				if err := os.MkdirAll(fullPath, 0o750); err != nil {
					t.Fatalf("Failed to create setup directory %s: %v", dirPath, err)
				}
			}

			env := &config.ViceEnv{
				ContextData: testContextDir,
				Context:     "test",
			}

			// Test the function
			result := IsFlotsamInitialized(env)

			if result != tt.expectedResult {
				t.Errorf("Expected %v but got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "dir-exists-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test existing directory
	if !dirExists(tempDir) {
		t.Error("Expected temp directory to exist")
	}

	// Test non-existing directory
	nonExistentDir := filepath.Join(tempDir, "non-existent")
	if dirExists(nonExistentDir) {
		t.Error("Expected non-existent directory to not exist")
	}

	// Test file vs directory
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if dirExists(testFile) {
		t.Error("Expected file to not be recognized as directory")
	}
}
