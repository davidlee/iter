package flotsam

import (
	"testing"
	"time"
)

func TestFlotsamReviewCreation(t *testing.T) {
	// AIDEV-NOTE: test review session creation and basic operations
	review := CreateFlotsamReview("work", "session-123")
	
	if review.Context != "work" {
		t.Errorf("Expected context 'work', got '%s'", review.Context)
	}
	
	if review.SessionID != "session-123" {
		t.Errorf("Expected session ID 'session-123', got '%s'", review.SessionID)
	}
	
	if review.Completed {
		t.Error("New review should not be completed")
	}
	
	if len(review.Items) != 0 {
		t.Errorf("New review should have 0 items, got %d", len(review.Items))
	}
}

func TestFlotsamReviewAddItem(t *testing.T) {
	review := CreateFlotsamReview("work", "session-123")
	
	previousSRS := &SRSData{
		Easiness:           2.5,
		ConsecutiveCorrect: 1,
		Due:               time.Now().AddDate(0, 0, -1).Unix(),
		TotalReviews:      1,
	}
	
	updatedSRS := &SRSData{
		Easiness:           2.7,
		ConsecutiveCorrect: 2,
		Due:               time.Now().AddDate(0, 0, 6).Unix(),
		TotalReviews:      2,
	}
	
	review.AddReviewItem("note1", CorrectHard, 30*time.Second, previousSRS, updatedSRS)
	
	if len(review.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(review.Items))
	}
	
	item := review.Items[0]
	if item.NoteID != "note1" {
		t.Errorf("Expected note ID 'note1', got '%s'", item.NoteID)
	}
	
	if item.Quality != CorrectHard {
		t.Errorf("Expected quality CorrectHard, got %v", item.Quality)
	}
	
	if item.ReviewTime != 30*time.Second {
		t.Errorf("Expected review time 30s, got %v", item.ReviewTime)
	}
	
	if item.PreviousSRSData.TotalReviews != 1 {
		t.Errorf("Expected previous reviews 1, got %d", item.PreviousSRSData.TotalReviews)
	}
	
	if item.UpdatedSRSData.TotalReviews != 2 {
		t.Errorf("Expected updated reviews 2, got %d", item.UpdatedSRSData.TotalReviews)
	}
}

func TestFlotsamReviewValidation(t *testing.T) {
	// AIDEV-NOTE: test review validation logic
	tests := []struct {
		name        string
		review      *FlotsamReview
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_review",
			review: &FlotsamReview{
				Context:   "work",
				SessionID: "session-1",
				Items: []FlotsamReviewItem{
					{
						NoteID:  "note1",
						Quality: CorrectHard,
						UpdatedSRSData: &SRSData{
							Easiness:     2.5,
							TotalReviews: 1,
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty_context",
			review: &FlotsamReview{
				Context: "",
				Items:   []FlotsamReviewItem{},
			},
			expectError: true,
			errorMsg:    "invalid context",
		},
		{
			name: "no_items",
			review: &FlotsamReview{
				Context: "work",
				Items:   []FlotsamReviewItem{},
			},
			expectError: true,
			errorMsg:    "at least one item",
		},
		{
			name: "duplicate_note_ids",
			review: &FlotsamReview{
				Context: "work",
				Items: []FlotsamReviewItem{
					{
						NoteID:  "note1",
						Quality: CorrectHard,
						UpdatedSRSData: &SRSData{Easiness: 2.5, TotalReviews: 1},
					},
					{
						NoteID:  "note1",
						Quality: CorrectEasy,
						UpdatedSRSData: &SRSData{Easiness: 2.6, TotalReviews: 1},
					},
				},
			},
			expectError: true,
			errorMsg:    "duplicate note ID",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.review.Validate()
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if err.Error() == "" {
					t.Error("Expected non-empty error message")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestFlotsamReviewItemValidation(t *testing.T) {
	tests := []struct {
		name        string
		item        FlotsamReviewItem
		expectError bool
	}{
		{
			name: "valid_item",
			item: FlotsamReviewItem{
				NoteID:  "note1",
				Quality: CorrectHard,
				UpdatedSRSData: &SRSData{
					Easiness:     2.5,
					TotalReviews: 1,
				},
			},
			expectError: false,
		},
		{
			name: "empty_note_id",
			item: FlotsamReviewItem{
				NoteID:  "",
				Quality: CorrectHard,
				UpdatedSRSData: &SRSData{Easiness: 2.5, TotalReviews: 1},
			},
			expectError: true,
		},
		{
			name: "invalid_quality",
			item: FlotsamReviewItem{
				NoteID:  "note1",
				Quality: Quality(10), // Invalid quality
				UpdatedSRSData: &SRSData{Easiness: 2.5, TotalReviews: 1},
			},
			expectError: true,
		},
		{
			name: "missing_updated_srs",
			item: FlotsamReviewItem{
				NoteID:         "note1",
				Quality:        CorrectHard,
				UpdatedSRSData: nil,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestFlotsamReviewStatistics(t *testing.T) {
	// AIDEV-NOTE: test review statistics calculation
	review := CreateFlotsamReview("test", "session-1")
	
	// Add some review items with different qualities
	review.AddReviewItem("note1", CorrectEasy, 20*time.Second, nil, &SRSData{TotalReviews: 1})
	review.AddReviewItem("note2", CorrectHard, 45*time.Second, nil, &SRSData{TotalReviews: 1})
	review.AddReviewItem("note3", IncorrectBlackout, 60*time.Second, nil, &SRSData{TotalReviews: 1})
	review.AddReviewItem("note4", CorrectEffort, 30*time.Second, nil, &SRSData{TotalReviews: 1})
	
	// Test counts
	if review.GetReviewCount() != 4 {
		t.Errorf("Expected 4 reviews, got %d", review.GetReviewCount())
	}
	
	if review.GetCorrectCount() != 3 {
		t.Errorf("Expected 3 correct, got %d", review.GetCorrectCount())
	}
	
	if review.GetIncorrectCount() != 1 {
		t.Errorf("Expected 1 incorrect, got %d", review.GetIncorrectCount())
	}
	
	// Test success rate
	expectedSuccessRate := 75.0 // 3/4 * 100
	if review.GetSuccessRate() != expectedSuccessRate {
		t.Errorf("Expected success rate %.1f, got %.1f", expectedSuccessRate, review.GetSuccessRate())
	}
	
	// Test average quality
	// Qualities: 6, 4, 1, 5 = 16/4 = 4.0
	expectedAvgQuality := 4.0
	if review.GetAverageQuality() != expectedAvgQuality {
		t.Errorf("Expected average quality %.1f, got %.1f", expectedAvgQuality, review.GetAverageQuality())
	}
	
	// Test timing
	expectedTotalTime := 155 * time.Second // 20+45+60+30
	if review.GetTotalReviewTime() != expectedTotalTime {
		t.Errorf("Expected total time %v, got %v", expectedTotalTime, review.GetTotalReviewTime())
	}
	
	expectedAvgTime := expectedTotalTime / 4
	if review.GetAverageReviewTime() != expectedAvgTime {
		t.Errorf("Expected average time %v, got %v", expectedAvgTime, review.GetAverageReviewTime())
	}
	
	// Test new cards
	if !review.HasNewCards() {
		t.Error("Review should have new cards (previous SRS data is nil)")
	}
	
	if review.GetNewCardCount() != 4 {
		t.Errorf("Expected 4 new cards, got %d", review.GetNewCardCount())
	}
}

func TestFlotsamDueCreation(t *testing.T) {
	due := CreateFlotsamDue("work")
	
	if due.Context != "work" {
		t.Errorf("Expected context 'work', got '%s'", due.Context)
	}
	
	if len(due.Items) != 0 {
		t.Errorf("New due list should have 0 items, got %d", len(due.Items))
	}
	
	if due.TotalDue != 0 {
		t.Errorf("New due list should have total 0, got %d", due.TotalDue)
	}
}

func TestFlotsamDueAddItem(t *testing.T) {
	due := CreateFlotsamDue("work")
	
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)
	
	// Add overdue item
	due.AddDueItem("note1", yesterday, false, "Overdue Note", "flashcard", 2.5, 3)
	
	// Add new card
	due.AddDueItem("note2", now, true, "New Note", "idea", 0, 0)
	
	// Add future item (shouldn't be overdue)
	due.AddDueItem("note3", tomorrow, false, "Future Note", "flashcard", 2.7, 5)
	
	if due.TotalDue != 3 {
		t.Errorf("Expected total due 3, got %d", due.TotalDue)
	}
	
	if due.Overdue != 1 {
		t.Errorf("Expected 1 overdue, got %d", due.Overdue)
	}
	
	if due.NewCards != 1 {
		t.Errorf("Expected 1 new card, got %d", due.NewCards)
	}
	
	// Check overdue item
	overdueItems := due.GetOverdueItems()
	if len(overdueItems) != 1 {
		t.Errorf("Expected 1 overdue item, got %d", len(overdueItems))
	}
	
	if overdueItems[0].NoteID != "note1" {
		t.Errorf("Expected overdue note 'note1', got '%s'", overdueItems[0].NoteID)
	}
	
	// Check new card items
	newCardItems := due.GetNewCardItems()
	if len(newCardItems) != 1 {
		t.Errorf("Expected 1 new card item, got %d", len(newCardItems))
	}
	
	if newCardItems[0].NoteID != "note2" {
		t.Errorf("Expected new card 'note2', got '%s'", newCardItems[0].NoteID)
	}
}

func TestFlotsamDueValidation(t *testing.T) {
	tests := []struct {
		name        string
		due         *FlotsamDue
		expectError bool
	}{
		{
			name: "valid_due",
			due: &FlotsamDue{
				Context:  "work",
				TotalDue: 1,
				Items: []FlotsamDueItem{
					{
						NoteID:      "note1",
						OverdueDays: 0,
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty_context",
			due: &FlotsamDue{
				Context:  "",
				TotalDue: 0,
				Items:    []FlotsamDueItem{},
			},
			expectError: true,
		},
		{
			name: "mismatched_count",
			due: &FlotsamDue{
				Context:  "work",
				TotalDue: 2,
				Items: []FlotsamDueItem{
					{NoteID: "note1", OverdueDays: 0},
				},
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.due.Validate()
			
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestFlotsamDueItemValidation(t *testing.T) {
	tests := []struct {
		name        string
		item        FlotsamDueItem
		expectError bool
	}{
		{
			name: "valid_item",
			item: FlotsamDueItem{
				NoteID:      "note1",
				OverdueDays: 5,
			},
			expectError: false,
		},
		{
			name: "empty_note_id",
			item: FlotsamDueItem{
				NoteID:      "",
				OverdueDays: 0,
			},
			expectError: true,
		},
		{
			name: "negative_overdue",
			item: FlotsamDueItem{
				NoteID:      "note1",
				OverdueDays: -1,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

func TestFlotsamDueSorting(t *testing.T) {
	// AIDEV-NOTE: test due list sorting functionality
	due := CreateFlotsamDue("test")
	
	now := time.Now()
	day1 := now.AddDate(0, 0, -3) // 3 days ago
	day2 := now.AddDate(0, 0, -1) // 1 day ago
	day3 := now                   // today
	
	due.AddDueItem("note1", day2, false, "Note 1", "flashcard", 2.5, 1)
	due.AddDueItem("note2", day1, false, "Note 2", "flashcard", 2.6, 2)
	due.AddDueItem("note3", day3, false, "Note 3", "flashcard", 2.4, 1)
	
	// Test sort by due date
	due.SortByDueDate()
	expectedOrder := []string{"note2", "note1", "note3"} // earliest first
	
	for i, expectedID := range expectedOrder {
		if due.Items[i].NoteID != expectedID {
			t.Errorf("Position %d: expected %s, got %s", i, expectedID, due.Items[i].NoteID)
		}
	}
	
	// Test sort by overdue days
	due.SortByOverdue()
	expectedOverdueOrder := []string{"note2", "note1", "note3"} // most overdue first
	
	for i, expectedID := range expectedOverdueOrder {
		if due.Items[i].NoteID != expectedID {
			t.Errorf("Overdue position %d: expected %s, got %s", i, expectedID, due.Items[i].NoteID)
		}
	}
}

func TestFlotsamReviewCompletion(t *testing.T) {
	review := CreateFlotsamReview("test", "session-1")
	
	// Add some items
	review.AddReviewItem("note1", CorrectEasy, 30*time.Second, nil, &SRSData{TotalReviews: 1})
	review.AddReviewItem("note2", CorrectHard, 45*time.Second, nil, &SRSData{TotalReviews: 1})
	
	if review.Completed {
		t.Error("Review should not be completed initially")
	}
	
	// Complete the review
	review.CompleteReview()
	
	if !review.Completed {
		t.Error("Review should be completed after calling CompleteReview")
	}
	
	expectedDuration := 75 * time.Second // 30 + 45
	if review.TotalDuration != expectedDuration {
		t.Errorf("Expected total duration %v, got %v", expectedDuration, review.TotalDuration)
	}
}

func TestFlotsamDueGetDueToday(t *testing.T) {
	due := CreateFlotsamDue("test")
	
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)
	
	// Add items with different due dates
	due.AddDueItem("overdue", yesterday, false, "Overdue", "flashcard", 2.5, 1)
	due.AddDueItem("today1", today, false, "Today 1", "flashcard", 2.6, 2)
	due.AddDueItem("today2", today.Add(2*time.Hour), false, "Today 2", "flashcard", 2.4, 1)
	due.AddDueItem("future", tomorrow, false, "Future", "flashcard", 2.7, 3)
	
	dueToday := due.GetDueToday()
	
	if len(dueToday) != 2 {
		t.Errorf("Expected 2 items due today, got %d", len(dueToday))
	}
	
	// Check that we got the right items
	todayIDs := make(map[string]bool)
	for _, item := range dueToday {
		todayIDs[item.NoteID] = true
	}
	
	if !todayIDs["today1"] || !todayIDs["today2"] {
		t.Error("Should include today1 and today2 in due today list")
	}
	
	if todayIDs["overdue"] || todayIDs["future"] {
		t.Error("Should not include overdue or future items in due today list")
	}
}