// Package debug provides centralized debug logging for the vice application.
// AIDEV-NOTE: T024-debug-flag; centralized debug logging system for systematic fault analysis
package debug

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger categories for different subsystems
const (
	CategoryModal     = "MODAL"
	CategoryField     = "FIELD"
	CategoryEntryMenu = "ENTRYMENU"
	CategoryGeneral   = "GENERAL"
)

// DebugLogger provides centralized debug logging with file output
//
//revive:disable-next-line:exported -- DebugLogger name follows singleton pattern
type DebugLogger struct {
	enabled   bool
	logFile   *os.File
	loggers   map[string]*log.Logger
	mu        sync.RWMutex
	startTime time.Time
}

var (
	instance *DebugLogger
	once     sync.Once
)

// GetInstance returns the singleton debug logger instance
func GetInstance() *DebugLogger {
	once.Do(func() {
		instance = &DebugLogger{
			enabled:   false,
			loggers:   make(map[string]*log.Logger),
			startTime: time.Now(),
		}
	})
	return instance
}

// Initialize sets up debug logging to a file in the specified directory
func (dl *DebugLogger) Initialize(configDir string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.enabled {
		return nil // Already initialized
	}

	// Create debug log file
	logPath := filepath.Join(configDir, "vice-debug.log")
	// #nosec G304 -- logPath is constructed from trusted configDir + literal filename
	file, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create debug log file %s: %w", logPath, err)
	}

	dl.logFile = file
	dl.enabled = true

	// Create category-specific loggers
	dl.createLogger(CategoryModal)
	dl.createLogger(CategoryField)
	dl.createLogger(CategoryEntryMenu)
	dl.createLogger(CategoryGeneral)

	// Log initialization
	dl.loggers[CategoryGeneral].Printf("Debug logging initialized at %s", logPath)
	dl.loggers[CategoryGeneral].Printf("Session started at %s", dl.startTime.Format(time.RFC3339))

	return nil
}

// createLogger creates a logger for a specific category
func (dl *DebugLogger) createLogger(category string) {
	if dl.logFile == nil {
		return
	}

	prefix := fmt.Sprintf("[%s] ", category)
	dl.loggers[category] = log.New(dl.logFile, prefix, log.LstdFlags|log.Lshortfile)
}

// IsEnabled returns whether debug logging is enabled
func (dl *DebugLogger) IsEnabled() bool {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	return dl.enabled
}

// Printf logs a formatted message for the specified category
func (dl *DebugLogger) Printf(category, format string, args ...interface{}) {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	if !dl.enabled {
		return
	}

	logger, exists := dl.loggers[category]
	if !exists {
		// Fall back to general logger
		logger = dl.loggers[CategoryGeneral]
	}

	if logger != nil {
		logger.Printf(format, args...)
	}
}

// Close closes the debug log file
func (dl *DebugLogger) Close() error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.logFile != nil {
		if dl.enabled {
			dl.loggers[CategoryGeneral].Printf("Debug logging session ended at %s", time.Now().Format(time.RFC3339))
			dl.loggers[CategoryGeneral].Printf("Session duration: %v", time.Since(dl.startTime))
		}

		err := dl.logFile.Close()
		dl.logFile = nil
		dl.enabled = false
		dl.loggers = make(map[string]*log.Logger)
		return err
	}
	return nil
}

// SetOutput sets the output destination for debug logging (for testing)
func (dl *DebugLogger) SetOutput(w io.Writer) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if dl.enabled {
		for _, logger := range dl.loggers {
			logger.SetOutput(w)
		}
	}
}

// Modal logs messages related to modal operations
func Modal(format string, args ...interface{}) {
	GetInstance().Printf(CategoryModal, format, args...)
}

// Field logs messages related to form field operations
func Field(format string, args ...interface{}) {
	GetInstance().Printf(CategoryField, format, args...)
}

// EntryMenu logs messages related to entry menu operations
func EntryMenu(format string, args ...interface{}) {
	GetInstance().Printf(CategoryEntryMenu, format, args...)
}

// General logs general debug messages
func General(format string, args ...interface{}) {
	GetInstance().Printf(CategoryGeneral, format, args...)
}
