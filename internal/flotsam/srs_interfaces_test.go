package flotsam

import (
	"testing"
	"time"
)

// Test that SM2Calculator implements the Algorithm interface
func TestSM2CalculatorImplementsAlgorithm(t *testing.T) {
	var _ Algorithm = (*SM2Calculator)(nil)
	
	calc := NewSM2Calculator()
	
	// Test basic interface compliance by calling methods
	if !calc.IsDue(nil) {
		t.Error("New cards should be due")
	}
	
	dueTime := calc.GetDueTime(nil)
	if dueTime.IsZero() {
		t.Error("Due time should not be zero")
	}
	
	interval := calc.GetNextInterval(nil)
	if interval != 0 {
		t.Error("New cards should have 0 interval")
	}
	
	// Test processing a review
	data, err := calc.ProcessReview(nil, CorrectHard)
	if err != nil {
		t.Fatalf("ProcessReview failed: %v", err)
	}
	
	if data.TotalReviews != 1 {
		t.Errorf("Expected 1 review, got %d", data.TotalReviews)
	}
}

// Mock implementations for testing interface compliance
type MockSRSStorage struct {
	data map[string]*SRSData
}

func NewMockSRSStorage() *MockSRSStorage {
	return &MockSRSStorage{
		data: make(map[string]*SRSData),
	}
}

func (m *MockSRSStorage) LoadSRSData(noteID string) (*SRSData, error) {
	data, exists := m.data[noteID]
	if !exists {
		return nil, ErrNoSRSData // New card
	}
	return data, nil
}

func (m *MockSRSStorage) SaveSRSData(noteID string, data *SRSData) error {
	m.data[noteID] = data
	return nil
}

func (m *MockSRSStorage) GetDueCards(t time.Time) ([]string, error) {
	var due []string
	calc := NewSM2CalculatorWithTime(t)
	
	for id, data := range m.data {
		if calc.IsDue(data) {
			due = append(due, id)
		}
	}
	return due, nil
}

func (m *MockSRSStorage) GetDueCardsByContext(_ string, t time.Time) ([]string, error) {
	// Simplified: ignore context for mock
	return m.GetDueCards(t)
}

func (m *MockSRSStorage) ListAllSRSCards() ([]string, error) {
	var cards []string
	for id := range m.data {
		cards = append(cards, id)
	}
	return cards, nil
}

func (m *MockSRSStorage) DeleteSRSData(noteID string) error {
	delete(m.data, noteID)
	return nil
}

// Test that MockSRSStorage implements the SRSStorage interface
func TestMockSRSStorageImplementsInterface(t *testing.T) {
	var _ SRSStorage = (*MockSRSStorage)(nil)
	
	storage := NewMockSRSStorage()
	
	// Test loading non-existent data
	data, err := storage.LoadSRSData("nonexistent")
	if err != ErrNoSRSData {
		t.Fatalf("Expected ErrNoSRSData, got: %v", err)
	}
	if data != nil {
		t.Error("Expected nil for non-existent card")
	}
	
	// Test saving and loading data
	srsData := &SRSData{
		Easiness:           2.5,
		ConsecutiveCorrect: 1,
		Due:               time.Now().Unix(),
		TotalReviews:      1,
	}
	
	err = storage.SaveSRSData("test1", srsData)
	if err != nil {
		t.Fatalf("SaveSRSData failed: %v", err)
	}
	
	loaded, err := storage.LoadSRSData("test1")
	if err != nil {
		t.Fatalf("LoadSRSData failed: %v", err)
	}
	
	if loaded.Easiness != srsData.Easiness {
		t.Errorf("Expected easiness %.2f, got %.2f", srsData.Easiness, loaded.Easiness)
	}
	
	// Test listing cards
	cards, err := storage.ListAllSRSCards()
	if err != nil {
		t.Fatalf("ListAllSRSCards failed: %v", err)
	}
	
	if len(cards) != 1 || cards[0] != "test1" {
		t.Errorf("Expected [test1], got %v", cards)
	}
	
	// Test deletion
	err = storage.DeleteSRSData("test1")
	if err != nil {
		t.Fatalf("DeleteSRSData failed: %v", err)
	}
	
	data, err = storage.LoadSRSData("test1")
	if err != ErrNoSRSData {
		t.Fatalf("Expected ErrNoSRSData after delete, got: %v", err)
	}
	if data != nil {
		t.Error("Expected nil after deletion")
	}
}

func TestSRSConfigDefaults(t *testing.T) {
	config := DefaultSRSConfig()
	
	if config.Algorithm != "sm2" {
		t.Errorf("Expected algorithm 'sm2', got '%s'", config.Algorithm)
	}
	
	if config.DefaultQuality != NoReview {
		t.Errorf("Expected default quality NoReview, got %v", config.DefaultQuality)
	}
	
	if config.MaxCardsPerSession != 50 {
		t.Errorf("Expected max cards 50, got %d", config.MaxCardsPerSession)
	}
	
	if !config.IncludeHistory {
		t.Error("Expected include history to be true")
	}
	
	if !config.AutoEnableForFlashcards {
		t.Error("Expected auto enable flashcards to be true")
	}
	
	if !config.ContextFiltering {
		t.Error("Expected context filtering to be true")
	}
}

func TestFlotsamNoteStructure(t *testing.T) {
	// AIDEV-NOTE: test flotsam note data structure and SRS integration
	note := &FlotsamNote{
		ID:       "test1",
		Title:    "Test Note",
		Type:     "flashcard",
		Tags:     []string{"learning", "test"},
		Created:  time.Now(),
		Modified: time.Now(),
		Body:     "This is a test note with **markdown** content",
		Links:    []string{"related-note", "another-note"},
		Backlinks: []string{"referring-note"},
		FilePath: "/path/to/test1.md",
		SRS: &SRSData{
			Easiness:           2.5,
			ConsecutiveCorrect: 0,
			Due:               time.Now().Unix(),
			TotalReviews:      0,
		},
	}
	
	// Verify structure
	if note.ID != "test1" {
		t.Errorf("Expected ID 'test1', got '%s'", note.ID)
	}
	
	if note.Type != "flashcard" {
		t.Errorf("Expected type 'flashcard', got '%s'", note.Type)
	}
	
	if len(note.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(note.Tags))
	}
	
	if note.SRS == nil {
		t.Error("Expected SRS data to be present")
	}
	
	if note.SRS.Easiness != 2.5 {
		t.Errorf("Expected easiness 2.5, got %.2f", note.SRS.Easiness)
	}
}

func TestSRSStatsStructure(t *testing.T) {
	stats := &SRSStats{
		TotalCards:       100,
		NewCards:         20,
		ReviewCards:      80,
		DueCards:         15,
		TotalReviews:     500,
		CorrectReviews:   400,
		IncorrectReviews: 100,
		SuccessRate:      80.0,
		AverageEasiness:  2.7,
		AverageInterval:  12.5,
		LongestInterval:  180,
		ShortestInterval: 1,
		DueToday:         15,
		DueTomorrow:      8,
		DueThisWeek:      25,
		DueThisMonth:     45,
		Overdue:          3,
	}
	
	// Verify calculations make sense
	if stats.TotalCards != stats.NewCards+stats.ReviewCards {
		t.Error("Total cards should equal new + review cards")
	}
	
	if stats.TotalReviews != stats.CorrectReviews+stats.IncorrectReviews {
		t.Error("Total reviews should equal correct + incorrect")
	}
	
	expectedSuccessRate := float64(stats.CorrectReviews) / float64(stats.TotalReviews) * 100
	if stats.SuccessRate != expectedSuccessRate {
		t.Errorf("Success rate mismatch: expected %.1f, got %.1f", expectedSuccessRate, stats.SuccessRate)
	}
}

func TestReviewSessionStructure(t *testing.T) {
	// AIDEV-NOTE: test review session management structures
	session := &ReviewSession{
		StartTime: time.Now(),
		Context:   "work",
		DueCards: []*FlotsamNote{
			{ID: "card1", Title: "Test Card 1"},
			{ID: "card2", Title: "Test Card 2"},
		},
		ReviewedCards: []*FlotsamNote{},
		CurrentIndex:  0,
		TotalCards:    2,
		CorrectCount:  0,
		ReviewedCount: 0,
		SessionStats: &SessionStats{
			Duration:       0,
			CardsReviewed:  0,
			CorrectAnswers: 0,
			IncorrectAnswers: 0,
			SuccessRate:    0,
			AverageTime:    0,
		},
	}
	
	if session.Context != "work" {
		t.Errorf("Expected context 'work', got '%s'", session.Context)
	}
	
	if len(session.DueCards) != 2 {
		t.Errorf("Expected 2 due cards, got %d", len(session.DueCards))
	}
	
	if session.TotalCards != 2 {
		t.Errorf("Expected total cards 2, got %d", session.TotalCards)
	}
	
	if session.SessionStats == nil {
		t.Error("Expected session stats to be initialized")
	}
}

func TestSRSErrors(t *testing.T) {
	// AIDEV-NOTE: test SRS error types and messages
	errors := []error{
		ErrNoteNotFound,
		ErrNoSRSData,
		ErrSRSAlreadyEnabled,
		ErrInvalidContext,
		ErrStorageFailure,
	}
	
	expectedMessages := []string{
		"note not found",
		"note has no SRS data",
		"SRS already enabled for this note",
		"invalid context",
		"storage operation failed",
	}
	
	for i, err := range errors {
		if err.Error() != expectedMessages[i] {
			t.Errorf("Error %d: expected '%s', got '%s'", i, expectedMessages[i], err.Error())
		}
	}
}

// Test interface type assertions to ensure all interfaces are properly defined
func TestInterfaceCompilation(t *testing.T) {
	// AIDEV-NOTE: compile-time interface validation
	
	// These should compile without error if interfaces are properly defined
	var _ Algorithm = (*SM2Calculator)(nil)
	var _ SRSStorage = (*MockSRSStorage)(nil)
	
	// Verify interface methods exist (will fail to compile if not)
	var algo Algorithm = NewSM2Calculator()
	var storage SRSStorage = NewMockSRSStorage()
	
	// Call interface methods to ensure they're callable
	_ = algo.IsDue(nil)
	_, _ = storage.LoadSRSData("test")
	
	// If we get here, all interfaces are properly defined
	t.Log("All interfaces compiled successfully")
}