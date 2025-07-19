// Package flotsam provides tag-based note detection and SRS integration.
// This file implements the vice:type:* tag hierarchy with zk delegation.
// AIDEV-NOTE: T041/4.3-tags; complete vice:type:* hierarchy replacing vice:srs redundancy
package flotsam

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/davidlee/vice/internal/config"
)

// Note type constants for vice:type:* tag hierarchy
const (
	TypeFlashcard = "flashcard" // Question/answer cards for SRS
	TypeIdea      = "idea"      // Free-form idea capture for SRS
	TypeScript    = "script"    // Executable scripts for SRS
	TypeLog       = "log"       // Journal entries for SRS
)

// GetFlashcardNotes returns all notes tagged with vice:type:flashcard.
func GetFlashcardNotes(env *config.ViceEnv) ([]string, error) {
	return GetNotesByType(env, TypeFlashcard)
}

// GetIdeaNotes returns all notes tagged with vice:type:idea.
func GetIdeaNotes(env *config.ViceEnv) ([]string, error) {
	return GetNotesByType(env, TypeIdea)
}

// GetScriptNotes returns all notes tagged with vice:type:script.
func GetScriptNotes(env *config.ViceEnv) ([]string, error) {
	return GetNotesByType(env, TypeScript)
}

// GetLogNotes returns all notes tagged with vice:type:log.
func GetLogNotes(env *config.ViceEnv) ([]string, error) {
	return GetNotesByType(env, TypeLog)
}

// GetAllViceNotes returns all notes with any vice:type:* tag.
// All returned notes are SRS-enabled by definition.
// AIDEV-NOTE: key function for bulk SRS operations, delegates to zk wildcard queries
func GetAllViceNotes(env *config.ViceEnv) ([]string, error) {
	if !env.IsZKAvailable() {
		log.Warn("ZK unavailable, cannot query vice-typed notes")
		return []string{}, fmt.Errorf("zk not available - install from https://github.com/zk-org/zk")
	}

	// Query for all vice:type:* tags
	// Note: zk supports wildcard tag queries
	notes, err := env.ZKList("--tag", "vice:type:*")
	if err != nil {
		log.Error("Failed to query all vice-typed notes", "error", err)
		return nil, fmt.Errorf("failed to query vice-typed notes: %w", err)
	}

	log.Debug("Found vice-typed notes", "count", len(notes))
	return notes, nil
}

// GetNotesByType returns all notes tagged with the specified vice:type.
func GetNotesByType(env *config.ViceEnv, noteType string) ([]string, error) {
	if !env.IsZKAvailable() {
		log.Warn("ZK unavailable, cannot query notes", "type", noteType)
		return []string{}, fmt.Errorf("zk not available - install from https://github.com/zk-org/zk")
	}

	// Construct vice:type:* tag
	tag := fmt.Sprintf("vice:type:%s", noteType)
	
	notes, err := env.ZKList("--tag", tag)
	if err != nil {
		log.Error("Failed to query notes by type", "type", noteType, "tag", tag, "error", err)
		return nil, fmt.Errorf("failed to query notes with tag %s: %w", tag, err)
	}

	log.Debug("Found notes by type", "type", noteType, "count", len(notes))
	return notes, nil
}

// ValidateNoteType checks if a note type is valid.
func ValidateNoteType(noteType string) error {
	validTypes := []string{TypeFlashcard, TypeIdea, TypeScript, TypeLog}
	
	for _, valid := range validTypes {
		if noteType == valid {
			return nil
		}
	}
	
	return fmt.Errorf("invalid note type '%s': must be one of %v", noteType, validTypes)
}

// GetViceTag returns the full vice:type tag for a note type.
func GetViceTag(noteType string) string {
	return fmt.Sprintf("vice:type:%s", noteType)
}

// ParseViceTag extracts the note type from a vice:type:* tag.
// Returns the type (e.g., "flashcard") and true if valid, or "", false if invalid.
func ParseViceTag(tag string) (string, bool) {
	prefix := "vice:type:"
	if !strings.HasPrefix(tag, prefix) {
		return "", false
	}
	
	noteType := strings.TrimPrefix(tag, prefix)
	if err := ValidateNoteType(noteType); err != nil {
		return "", false
	}
	
	return noteType, true
}

// IsViceTag returns true if the tag is a valid vice:type:* tag.
func IsViceTag(tag string) bool {
	_, valid := ParseViceTag(tag)
	return valid
}

// EnrichWithSRSData combines note paths with SRS database information.
// Returns a map of note path to SRS data, with warnings for missing entries.
func EnrichWithSRSData(env *config.ViceEnv, notes []string) (map[string]*SRSData, error) {
	if len(notes) == 0 {
		return make(map[string]*SRSData), nil
	}

	// TODO: Implement bulk SRS database query
	// This requires extending internal/srs/database.go with bulk query methods
	// For now, return empty map to establish the interface
	
	log.Debug("Enriching notes with SRS data", "note_count", len(notes))
	
	srsData := make(map[string]*SRSData)
	
	// Placeholder: would bulk query SRS database here
	// srsResults := database.GetSRSDataBulk(notes)
	
	// Log warnings for notes missing SRS data
	for _, notePath := range notes {
		if srsData[notePath] == nil {
			log.Warn("Vice-typed note missing from SRS database", "path", notePath)
		}
	}
	
	return srsData, nil
}

// ValidateSRSConsistency checks that all vice-typed notes have SRS database entries.
// Logs warnings for inconsistencies but does not return errors.
func ValidateSRSConsistency(env *config.ViceEnv, notes []string) error {
	log.Debug("Validating SRS consistency", "note_count", len(notes))
	
	srsData, err := EnrichWithSRSData(env, notes)
	if err != nil {
		return fmt.Errorf("failed to check SRS consistency: %w", err)
	}
	
	missingCount := 0
	for _, notePath := range notes {
		if srsData[notePath] == nil {
			missingCount++
		}
	}
	
	if missingCount > 0 {
		log.Warn("SRS consistency issues detected", 
			"total_notes", len(notes), 
			"missing_srs_entries", missingCount)
	} else {
		log.Debug("SRS consistency validation passed", "note_count", len(notes))
	}
	
	return nil
}