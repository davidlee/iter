package checklist

import (
	"strings"
	"testing"

	"davidlee/vice/internal/models"
)

func TestCompletionModel_getSectionProgress(t *testing.T) {
	// Create test checklist with sections
	checklist := &models.Checklist{
		ID:    "test",
		Title: "Test Checklist",
		Items: []string{
			"# section one",
			"item 1",
			"item 2",
			"# section two",
			"item 3",
			"item 4",
			"item 5",
			"# empty section",
			"# section three",
			"item 6",
		},
	}

	model := NewCompletionModel(checklist)

	// Select some items
	model.selected[1] = struct{}{} // item 1 (section one)
	model.selected[4] = struct{}{} // item 3 (section two)
	model.selected[5] = struct{}{} // item 4 (section two)
	model.selected[9] = struct{}{} // item 6 (section three)

	tests := []struct {
		name          string
		headingIndex  int
		wantCompleted int
		wantTotal     int
	}{
		{
			name:          "section one progress",
			headingIndex:  0, // "# section one"
			wantCompleted: 1, // item 1 selected
			wantTotal:     2, // item 1, item 2
		},
		{
			name:          "section two progress",
			headingIndex:  3, // "# section two"
			wantCompleted: 2, // item 3, item 4 selected
			wantTotal:     3, // item 3, item 4, item 5
		},
		{
			name:          "empty section",
			headingIndex:  7, // "# empty section"
			wantCompleted: 0,
			wantTotal:     0,
		},
		{
			name:          "section three progress",
			headingIndex:  8, // "# section three"
			wantCompleted: 1, // item 6 selected
			wantTotal:     1, // item 6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completed, total := model.getSectionProgress(tt.headingIndex)
			if completed != tt.wantCompleted {
				t.Errorf("getSectionProgress() completed = %v, want %v", completed, tt.wantCompleted)
			}
			if total != tt.wantTotal {
				t.Errorf("getSectionProgress() total = %v, want %v", total, tt.wantTotal)
			}
		})
	}
}

func TestCompletionModel_ViewWithProgressIndicators(t *testing.T) {
	// Create test checklist
	checklist := &models.Checklist{
		ID:    "test",
		Title: "Test Checklist",
		Items: []string{
			"# clean station",
			"clear desk",
			"clear inbox",
			"# digital inputs",
			"process emails",
			"check phone",
		},
	}

	model := NewCompletionModel(checklist)

	// Select first item in each section
	model.selected[1] = struct{}{} // clear desk
	model.selected[4] = struct{}{} // process emails

	view := model.View()

	// Check that progress indicators appear in headings
	if !strings.Contains(view, "clean station (1/2)") {
		t.Errorf("Expected 'clean station (1/2)' in view, got: %s", view)
	}

	if !strings.Contains(view, "digital inputs (1/2)") {
		t.Errorf("Expected 'digital inputs (1/2)' in view, got: %s", view)
	}

	// Check that overall progress information appears
	if !strings.Contains(view, "Completed: 2/4 items (50%)") {
		t.Errorf("Expected 'Completed: 2/4 items (50%%)' in view, got: %s", view)
	}
}

func TestCompletionModel_getSectionProgressEdgeCases(t *testing.T) {
	checklist := &models.Checklist{
		ID:    "test",
		Title: "Test Checklist",
		Items: []string{
			"item without heading",
			"# heading",
			"item after heading",
		},
	}

	model := NewCompletionModel(checklist)

	tests := []struct {
		name          string
		headingIndex  int
		wantCompleted int
		wantTotal     int
	}{
		{
			name:          "invalid index",
			headingIndex:  -1,
			wantCompleted: 0,
			wantTotal:     0,
		},
		{
			name:          "index beyond items",
			headingIndex:  10,
			wantCompleted: 0,
			wantTotal:     0,
		},
		{
			name:          "non-heading item",
			headingIndex:  0, // "item without heading"
			wantCompleted: 0,
			wantTotal:     0,
		},
		{
			name:          "valid heading",
			headingIndex:  1, // "# heading"
			wantCompleted: 0,
			wantTotal:     1, // "item after heading"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completed, total := model.getSectionProgress(tt.headingIndex)
			if completed != tt.wantCompleted {
				t.Errorf("getSectionProgress() completed = %v, want %v", completed, tt.wantCompleted)
			}
			if total != tt.wantTotal {
				t.Errorf("getSectionProgress() total = %v, want %v", total, tt.wantTotal)
			}
		})
	}
}
