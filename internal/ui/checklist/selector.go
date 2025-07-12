package checklist

import (
	"fmt"

	"github.com/charmbracelet/huh"

	"davidlee/iter/internal/models"
)

// Selector provides a menu interface for selecting checklists.
type Selector struct {
	checklists []models.Checklist
}

// NewSelector creates a new checklist selector.
func NewSelector(checklists []models.Checklist) *Selector {
	return &Selector{checklists: checklists}
}

// SelectChecklist displays a menu to select a checklist and returns the selected one.
func (s *Selector) SelectChecklist() (*models.Checklist, error) {
	if len(s.checklists) == 0 {
		return nil, fmt.Errorf("no checklists available")
	}

	// Prepare options for the selector
	options := make([]huh.Option[string], len(s.checklists))
	for i, checklist := range s.checklists {
		// Create a descriptive label
		label := checklist.Title
		if label == "" {
			label = checklist.ID
		}

		// Add item count info
		itemCount := checklist.GetTotalItemCount()

		// Include item count and description in the label
		fullLabel := label
		if checklist.Description != "" {
			fullLabel += fmt.Sprintf(" - %s", checklist.Description)
		}
		fullLabel += fmt.Sprintf(" (%d items)", itemCount)

		options[i] = huh.NewOption(fullLabel, checklist.ID)
	}

	var selectedID string

	// Create the selection form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a checklist to complete").
				Description("Choose from your available checklists").
				Options(options...).
				Value(&selectedID),
		),
	)

	// Run the form
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("selection error: %w", err)
	}

	// Find and return the selected checklist
	for i := range s.checklists {
		if s.checklists[i].ID == selectedID {
			return &s.checklists[i], nil
		}
	}

	return nil, fmt.Errorf("selected checklist not found")
}
