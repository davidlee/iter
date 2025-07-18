package checklist

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/davidlee/vice/internal/models"
)

// EditorConfig holds configuration for the checklist editor.
type EditorConfig struct {
	ChecklistID   string
	Title         string
	Description   string
	ExistingItems []string // For edit mode
	IsEdit        bool
	GenerateID    bool // Generate ID from title if ChecklistID is empty
}

// Editor provides a simple multiline text interface for creating/editing checklists.
type Editor struct {
	config EditorConfig
}

// NewEditor creates a new checklist editor.
func NewEditor(config EditorConfig) *Editor {
	return &Editor{config: config}
}

// Run displays the editor UI and returns the configured checklist.
func (e *Editor) Run() (*models.Checklist, error) {
	// Prepare initial content for edit mode
	var initialContent string
	if e.config.IsEdit && len(e.config.ExistingItems) > 0 {
		initialContent = strings.Join(e.config.ExistingItems, "\n")
	}

	// Form fields
	var title, description, itemsText string

	// Set initial values
	title = e.config.Title
	description = e.config.Description
	itemsText = initialContent

	// Create the form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Checklist Title").
				Description("Enter a descriptive title for this checklist").
				Value(&title).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("title is required")
					}
					return nil
				}),

			huh.NewInput().
				Title("Description").
				Description("Optional description of what this checklist is for").
				Value(&description),

			huh.NewText().
				Title("Checklist Items").
				Description("Enter checklist items, one per line. Use '# ' prefix for headings.").
				Placeholder(e.getPlaceholderText()).
				Value(&itemsText).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("at least one item is required")
					}
					return nil
				}),
		),
	)

	// Run the form
	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("form error: %w", err)
	}

	// Parse the items
	items := e.parseItems(itemsText)
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}

	// Generate ID from title if needed
	checklistID := e.config.ChecklistID
	if e.config.GenerateID && checklistID == "" {
		checklistID = generateChecklistIDFromTitle(strings.TrimSpace(title))
	}

	// Create the checklist model
	checklist := &models.Checklist{
		ID:           checklistID,
		Title:        strings.TrimSpace(title),
		Description:  strings.TrimSpace(description),
		Items:        items,
		CreatedDate:  time.Now().Format("2006-01-02"),
		ModifiedDate: time.Now().Format("2006-01-02"),
	}

	// For edit mode, preserve original creation date if available
	// Note: The ModifiedDate will be updated, but CreatedDate should be preserved
	// This will be handled by the calling code that has access to the original checklist

	return checklist, nil
}

// parseItems parses the multiline text input into a slice of checklist items.
func (e *Editor) parseItems(text string) []string {
	lines := strings.Split(text, "\n")
	var items []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}

	return items
}

// getPlaceholderText returns example text for the items field.
func (e *Editor) getPlaceholderText() string {
	return `# Morning Setup
check email
clear desk

# Daily Planning  
review calendar
set priorities`
}

// ValidateChecklistID checks if a checklist ID is valid for creation.
func ValidateChecklistID(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("checklist ID is required")
	}

	// Use the same validation as habit IDs
	trimmed := strings.TrimSpace(id)
	for _, char := range trimmed {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '_' {
			return fmt.Errorf("checklist ID must contain only lowercase letters, numbers, and underscores")
		}
	}

	return nil
}

// generateChecklistIDFromTitle creates a valid ID from a checklist title.
// Uses the same logic as habit ID generation for consistency.
// AIDEV-NOTE: id-consistency; mirrors internal/models/habit.go:generateIDFromTitle
func generateChecklistIDFromTitle(title string) string {
	// Convert to lowercase
	id := strings.ToLower(title)

	// Replace spaces and special characters with underscores
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	id = reg.ReplaceAllString(id, "_")

	// Remove consecutive underscores
	reg = regexp.MustCompile(`_+`)
	id = reg.ReplaceAllString(id, "_")

	// Trim leading/trailing underscores
	id = strings.Trim(id, "_")

	// Ensure it's not empty
	if id == "" {
		id = "unnamed_checklist"
	}

	return id
}
