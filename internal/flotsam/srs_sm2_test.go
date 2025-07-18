package flotsam

import (
	"math"
	"testing"
	"time"
)

func TestQualityValidation(t *testing.T) {
	validQualities := []Quality{
		NoReview, IncorrectBlackout, IncorrectFamiliar, IncorrectEasy,
		CorrectHard, CorrectEffort, CorrectEasy,
	}

	for _, q := range validQualities {
		if err := q.Validate(); err != nil {
			t.Errorf("Quality %d should be valid, got error: %v", q, err)
		}
	}

	invalidQualities := []Quality{Quality(-1), Quality(7), Quality(100)}
	for _, q := range invalidQualities {
		if err := q.Validate(); err == nil {
			t.Errorf("Quality %d should be invalid, but validation passed", q)
		}
	}
}

func TestQualityIsCorrect(t *testing.T) {
	incorrectQualities := []Quality{
		NoReview, IncorrectBlackout, IncorrectFamiliar, IncorrectEasy,
	}

	for _, q := range incorrectQualities {
		if q.IsCorrect() {
			t.Errorf("Quality %d should not be correct", q)
		}
	}

	correctQualities := []Quality{CorrectHard, CorrectEffort, CorrectEasy}
	for _, q := range correctQualities {
		if !q.IsCorrect() {
			t.Errorf("Quality %d should be correct", q)
		}
	}
}

func TestSM2CalculatorNewCard(t *testing.T) {
	// AIDEV-NOTE: test new card creation with various quality ratings
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	calc := NewSM2CalculatorWithTime(fixedTime)

	tests := []struct {
		name             string
		quality          Quality
		expectedEasiness float64
		expectedCorrect  int
		expectedDue      int64
	}{
		{
			name:             "no_review",
			quality:          NoReview,
			expectedEasiness: DefaultEasiness,
			expectedCorrect:  0,
			expectedDue:      fixedTime.AddDate(0, 0, 1).Unix(),
		},
		{
			name:             "first_correct_hard",
			quality:          CorrectHard,
			expectedEasiness: 2.72, // Calculated: 2.5 + (-0.8) + (0.28 * 3) + (0.02 * 9)
			expectedCorrect:  1,
			expectedDue:      fixedTime.AddDate(0, 0, 1).Unix(),
		},
		{
			name:             "first_correct_easy",
			quality:          CorrectEasy,
			expectedEasiness: 3.6, // Calculated: 2.5 + (-0.8) + (0.28 * 5) + (0.02 * 25)
			expectedCorrect:  1,
			expectedDue:      fixedTime.AddDate(0, 0, 1).Unix(),
		},
		{
			name:             "first_incorrect",
			quality:          IncorrectBlackout,
			expectedEasiness: 1.7, // Calculated: 2.5 + (-0.8) + (0.28 * 0) + (0.02 * 0)
			expectedCorrect:  0,
			expectedDue:      fixedTime.AddDate(0, 0, 1).Unix(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := calc.ProcessReview(nil, tt.quality)
			if err != nil {
				t.Fatalf("ProcessReview failed: %v", err)
			}

			if math.Abs(data.Easiness-tt.expectedEasiness) > 0.01 {
				t.Errorf("Expected easiness %.2f, got %.2f", tt.expectedEasiness, data.Easiness)
			}

			if data.ConsecutiveCorrect != tt.expectedCorrect {
				t.Errorf("Expected consecutive correct %d, got %d", tt.expectedCorrect, data.ConsecutiveCorrect)
			}

			if data.Due != tt.expectedDue {
				t.Errorf("Expected due %d, got %d", tt.expectedDue, data.Due)
			}

			if data.TotalReviews != 1 {
				t.Errorf("Expected total reviews 1, got %d", data.TotalReviews)
			}

			if len(data.ReviewHistory) != 1 {
				t.Errorf("Expected 1 review in history, got %d", len(data.ReviewHistory))
			}
		})
	}
}

func TestSM2CalculatorUpdateCard(t *testing.T) {
	// AIDEV-NOTE: test card updates with various scenarios
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	calc := NewSM2CalculatorWithTime(fixedTime)

	// Start with a card that has been reviewed once correctly
	initialData := &SRSData{
		Easiness:           2.5,
		ConsecutiveCorrect: 1,
		Due:                fixedTime.AddDate(0, 0, -1).Unix(), // Due yesterday
		TotalReviews:       1,
		ReviewHistory:      []ReviewRecord{{Timestamp: fixedTime.AddDate(0, 0, -2).Unix(), Quality: CorrectHard}},
	}

	tests := []struct {
		name            string
		quality         Quality
		expectedCorrect int
		expectedDueDays int // Days from now
		minEasiness     float64
		maxEasiness     float64
	}{
		{
			name:            "second_correct_hard",
			quality:         CorrectHard,
			expectedCorrect: 2,
			expectedDueDays: 6, // DueDateStartDays = 6
			minEasiness:     2.7,
			maxEasiness:     2.8,
		},
		{
			name:            "second_correct_easy",
			quality:         CorrectEasy,
			expectedCorrect: 2,
			expectedDueDays: 6,
			minEasiness:     3.5,
			maxEasiness:     3.7,
		},
		{
			name:            "second_incorrect",
			quality:         IncorrectBlackout,
			expectedCorrect: 0,
			expectedDueDays: 1, // Reset to tomorrow
			minEasiness:     1.6,
			maxEasiness:     1.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := calc.ProcessReview(initialData, tt.quality)
			if err != nil {
				t.Fatalf("ProcessReview failed: %v", err)
			}

			if data.ConsecutiveCorrect != tt.expectedCorrect {
				t.Errorf("Expected consecutive correct %d, got %d", tt.expectedCorrect, data.ConsecutiveCorrect)
			}

			expectedDue := fixedTime.AddDate(0, 0, tt.expectedDueDays).Unix()
			if data.Due != expectedDue {
				t.Errorf("Expected due %d (%s), got %d (%s)",
					expectedDue, time.Unix(expectedDue, 0),
					data.Due, time.Unix(data.Due, 0))
			}

			if data.Easiness < tt.minEasiness || data.Easiness > tt.maxEasiness {
				t.Errorf("Expected easiness between %.2f and %.2f, got %.2f",
					tt.minEasiness, tt.maxEasiness, data.Easiness)
			}

			if data.TotalReviews != 2 {
				t.Errorf("Expected total reviews 2, got %d", data.TotalReviews)
			}

			if len(data.ReviewHistory) != 2 {
				t.Errorf("Expected 2 reviews in history, got %d", len(data.ReviewHistory))
			}
		})
	}
}

func TestSM2CalculatorIntervalGrowth(t *testing.T) {
	// AIDEV-NOTE: test exponential growth of intervals with consecutive correct answers
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	calc := NewSM2CalculatorWithTime(fixedTime)

	// Simulate a card being reviewed correctly multiple times
	var data *SRSData
	var err error

	expectedIntervals := []int{1, 6} // First correct: 1 day, Second: 6 days

	for i := 0; i < 2; i++ {
		data, err = calc.ProcessReview(data, CorrectEffort)
		if err != nil {
			t.Fatalf("ProcessReview %d failed: %v", i+1, err)
		}

		interval := calc.GetNextInterval(data)
		if interval != expectedIntervals[i] {
			t.Errorf("Review %d: expected interval %d days, got %d days",
				i+1, expectedIntervals[i], interval)
		}
	}

	// Third and subsequent reviews should follow exponential growth
	// Formula: 6 * easiness^(consecutive-1)
	for i := 2; i < 5; i++ {
		// Move time forward to the due date
		calc.now = time.Unix(data.Due, 0)

		data, err = calc.ProcessReview(data, CorrectEffort)
		if err != nil {
			t.Fatalf("ProcessReview %d failed: %v", i+1, err)
		}

		// Verify interval grows exponentially
		interval := calc.GetNextInterval(data)
		if interval <= expectedIntervals[len(expectedIntervals)-1] {
			t.Errorf("Review %d: interval %d should be greater than previous %d",
				i+1, interval, expectedIntervals[len(expectedIntervals)-1])
		}

		expectedIntervals = append(expectedIntervals, interval)
	}
}

func TestSM2CalculatorIsDue(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	calc := NewSM2CalculatorWithTime(fixedTime)

	tests := []struct {
		name     string
		data     *SRSData
		expected bool
	}{
		{
			name:     "nil_data_new_card",
			data:     nil,
			expected: true,
		},
		{
			name: "due_yesterday",
			data: &SRSData{
				Due: fixedTime.AddDate(0, 0, -1).Unix(),
			},
			expected: true,
		},
		{
			name: "due_now",
			data: &SRSData{
				Due: fixedTime.Unix(),
			},
			expected: true,
		},
		{
			name: "due_tomorrow",
			data: &SRSData{
				Due: fixedTime.AddDate(0, 0, 1).Unix(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.IsDue(tt.data)
			if result != tt.expected {
				t.Errorf("Expected IsDue %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSM2CalculatorGetDueTime(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	calc := NewSM2CalculatorWithTime(fixedTime)

	dueTime := fixedTime.AddDate(0, 0, 5)
	data := &SRSData{Due: dueTime.Unix()}

	result := calc.GetDueTime(data)
	if !result.Equal(dueTime) {
		t.Errorf("Expected due time %v, got %v", dueTime, result)
	}

	// Test nil data
	result = calc.GetDueTime(nil)
	if !result.Equal(fixedTime) {
		t.Errorf("Expected due time for nil data %v, got %v", fixedTime, result)
	}
}

func TestSRSDataSerialization(t *testing.T) {
	// AIDEV-NOTE: test JSON serialization for frontmatter storage
	original := &SRSData{
		Easiness:           2.7,
		ConsecutiveCorrect: 3,
		Due:                1640995200,
		TotalReviews:       5,
		ReviewHistory: []ReviewRecord{
			{Timestamp: 1640995100, Quality: CorrectHard},
			{Timestamp: 1640995000, Quality: CorrectEasy},
		},
	}

	// Test serialization
	jsonData, err := SerializeSRSData(original)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	// Test deserialization
	restored, err := DeserializeSRSData(jsonData)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	// Verify all fields match
	if restored.Easiness != original.Easiness {
		t.Errorf("Easiness mismatch: expected %.2f, got %.2f", original.Easiness, restored.Easiness)
	}

	if restored.ConsecutiveCorrect != original.ConsecutiveCorrect {
		t.Errorf("ConsecutiveCorrect mismatch: expected %d, got %d", original.ConsecutiveCorrect, restored.ConsecutiveCorrect)
	}

	if restored.Due != original.Due {
		t.Errorf("Due mismatch: expected %d, got %d", original.Due, restored.Due)
	}

	if restored.TotalReviews != original.TotalReviews {
		t.Errorf("TotalReviews mismatch: expected %d, got %d", original.TotalReviews, restored.TotalReviews)
	}

	if len(restored.ReviewHistory) != len(original.ReviewHistory) {
		t.Errorf("ReviewHistory length mismatch: expected %d, got %d", len(original.ReviewHistory), len(restored.ReviewHistory))
	}

	// Test nil data
	jsonData, err = SerializeSRSData(nil)
	if err != nil {
		t.Fatalf("Nil serialization failed: %v", err)
	}

	if len(jsonData) != 0 {
		t.Errorf("Expected empty JSON for nil data, got %v", jsonData)
	}

	restored, err = DeserializeSRSData(nil)
	if err != ErrEmptyData {
		t.Fatalf("Expected ErrEmptyData for nil JSON, got: %v", err)
	}

	if restored != nil {
		t.Errorf("Expected nil data for nil JSON, got %v", restored)
	}
}

func TestSM2CalculatorEasinessLimits(t *testing.T) {
	// AIDEV-NOTE: test easiness factor boundaries and clamping
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	calc := NewSM2CalculatorWithTime(fixedTime)

	// Test minimum easiness clamping with very poor performance
	data := &SRSData{
		Easiness:           MinEasiness + 0.1,
		ConsecutiveCorrect: 1,
		TotalReviews:       1,
	}

	// Multiple incorrect answers should not reduce easiness below minimum
	for i := 0; i < 5; i++ {
		data, _ = calc.ProcessReview(data, IncorrectBlackout)
		if data.Easiness < MinEasiness {
			t.Errorf("Easiness %.2f below minimum %.2f after %d incorrect reviews",
				data.Easiness, MinEasiness, i+1)
		}
	}

	// Test that easiness can increase with good performance
	initialEasiness := data.Easiness
	data, _ = calc.ProcessReview(data, CorrectEasy)

	if data.Easiness <= initialEasiness {
		t.Errorf("Easiness should increase with correct answer: %.2f -> %.2f",
			initialEasiness, data.Easiness)
	}
}

func TestSM2CalculatorInvalidQuality(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	calc := NewSM2CalculatorWithTime(fixedTime)

	invalidQualities := []Quality{Quality(-1), Quality(7), Quality(100)}

	for _, q := range invalidQualities {
		_, err := calc.ProcessReview(nil, q)
		if err == nil {
			t.Errorf("Expected error for invalid quality %d, but got none", q)
		}
	}
}
