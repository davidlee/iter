package flotsam

import (
	"testing"

	"github.com/davidlee/vice/internal/config"
	"github.com/davidlee/vice/internal/zk"
)

func TestValidateNoteType(t *testing.T) {
	tests := []struct {
		name     string
		noteType string
		wantErr  bool
	}{
		{"valid flashcard", TypeFlashcard, false},
		{"valid idea", TypeIdea, false},
		{"valid script", TypeScript, false},
		{"valid log", TypeLog, false},
		{"invalid type", "invalid", true},
		{"empty string", "", true},
		{"case sensitive", "Flashcard", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNoteType(tt.noteType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNoteType(%q) error = %v, wantErr %v", tt.noteType, err, tt.wantErr)
			}
		})
	}
}

func TestGetViceTag(t *testing.T) {
	tests := []struct {
		noteType string
		want     string
	}{
		{TypeFlashcard, "vice:type:flashcard"},
		{TypeIdea, "vice:type:idea"},
		{TypeScript, "vice:type:script"},
		{TypeLog, "vice:type:log"},
	}

	for _, tt := range tests {
		t.Run(tt.noteType, func(t *testing.T) {
			got := GetViceTag(tt.noteType)
			if got != tt.want {
				t.Errorf("GetViceTag(%q) = %q, want %q", tt.noteType, got, tt.want)
			}
		})
	}
}

func TestParseViceTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		wantType string
		wantOK   bool
	}{
		{"valid flashcard tag", "vice:type:flashcard", "flashcard", true},
		{"valid idea tag", "vice:type:idea", "idea", true},
		{"valid script tag", "vice:type:script", "script", true},
		{"valid log tag", "vice:type:log", "log", true},
		{"invalid prefix", "vice:flashcard", "", false},
		{"invalid type", "vice:type:invalid", "", false},
		{"no prefix", "flashcard", "", false},
		{"empty string", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotOK := ParseViceTag(tt.tag)
			if gotType != tt.wantType || gotOK != tt.wantOK {
				t.Errorf("ParseViceTag(%q) = (%q, %v), want (%q, %v)",
					tt.tag, gotType, gotOK, tt.wantType, tt.wantOK)
			}
		})
	}
}

func TestIsViceTag(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{"vice:type:flashcard", true},
		{"vice:type:idea", true},
		{"vice:type:script", true},
		{"vice:type:log", true},
		{"vice:type:invalid", false},
		{"vice:flashcard", false},
		{"flashcard", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := IsViceTag(tt.tag)
			if got != tt.want {
				t.Errorf("IsViceTag(%q) = %v, want %v", tt.tag, got, tt.want)
			}
		})
	}
}

func TestGetNotesByType_ZKUnavailable(t *testing.T) {
	env := &config.ViceEnv{
		ZK: nil, // No ZK available
	}

	notes, err := GetNotesByType(env, TypeFlashcard)
	if err == nil {
		t.Error("GetNotesByType should return error when ZK unavailable")
	}

	if len(notes) != 0 {
		t.Errorf("GetNotesByType should return empty slice when ZK unavailable, got %d notes", len(notes))
	}

	expectedMsg := "zk not available - install from https://github.com/zk-org/zk"
	if err.Error() != expectedMsg {
		t.Errorf("GetNotesByType error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestGetAllViceNotes_ZKUnavailable(t *testing.T) {
	env := &config.ViceEnv{
		ZK: nil, // No ZK available
	}

	notes, err := GetAllViceNotes(env)
	if err == nil {
		t.Error("GetAllViceNotes should return error when ZK unavailable")
	}

	if len(notes) != 0 {
		t.Errorf("GetAllViceNotes should return empty slice when ZK unavailable, got %d notes", len(notes))
	}
}

func TestGetFlashcardNotes_ZKUnavailable(t *testing.T) {
	env := &config.ViceEnv{
		ZK: &zk.ZKExecutable{}, // Create unavailable ZK instance
	}

	notes, err := GetFlashcardNotes(env)
	if err == nil {
		t.Error("GetFlashcardNotes should return error when ZK unavailable")
	}

	if len(notes) != 0 {
		t.Errorf("GetFlashcardNotes should return empty slice when ZK unavailable, got %d notes", len(notes))
	}
}

func TestGetIdeaNotes_ZKUnavailable(t *testing.T) {
	env := &config.ViceEnv{
		ZK: &zk.ZKExecutable{}, // Create unavailable ZK instance
	}

	notes, err := GetIdeaNotes(env)
	if err == nil {
		t.Error("GetIdeaNotes should return error when ZK unavailable")
	}

	if len(notes) != 0 {
		t.Errorf("GetIdeaNotes should return empty slice when ZK unavailable, got %d notes", len(notes))
	}
}

func TestEnrichWithSRSData_EmptyNotes(t *testing.T) {
	env := &config.ViceEnv{}

	srsData, err := EnrichWithSRSData(env, []string{})
	if err != nil {
		t.Errorf("EnrichWithSRSData should not error for empty notes, got: %v", err)
	}

	if len(srsData) != 0 {
		t.Errorf("EnrichWithSRSData should return empty map for empty notes, got %d entries", len(srsData))
	}
}

func TestValidateSRSConsistency_EmptyNotes(t *testing.T) {
	env := &config.ViceEnv{}

	err := ValidateSRSConsistency(env, []string{})
	if err != nil {
		t.Errorf("ValidateSRSConsistency should not error for empty notes, got: %v", err)
	}
}

func TestConstantsDefinition(t *testing.T) {
	// Verify constants are properly defined
	expectedTypes := map[string]string{
		TypeFlashcard: "flashcard",
		TypeIdea:      "idea",
		TypeScript:    "script",
		TypeLog:       "log",
	}

	for constant, expected := range expectedTypes {
		if constant != expected {
			t.Errorf("Constant mismatch: %s = %q, want %q", constant, constant, expected)
		}
	}
}