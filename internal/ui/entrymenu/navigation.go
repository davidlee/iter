package entrymenu

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"

	"davidlee/vice/internal/models"
)

// NavigationHelper provides enhanced navigation capabilities for the entry menu.
// AIDEV-NOTE: navigation-helper; centralizes smart navigation logic for entry workflow
type NavigationHelper struct{}

// NewNavigationHelper creates a new navigation helper.
func NewNavigationHelper() *NavigationHelper {
	return &NavigationHelper{}
}

// FindNextIncompleteGoal finds the next goal that hasn't been entered yet.
func (n *NavigationHelper) FindNextIncompleteGoal(goals []models.Goal, entries map[string]models.GoalEntry, currentIndex int) int {
	// Start from the next position after current
	for i := currentIndex + 1; i < len(goals); i++ {
		if _, hasEntry := entries[goals[i].ID]; !hasEntry {
			return i
		}
	}
	
	// Wrap around to the beginning
	for i := 0; i <= currentIndex; i++ {
		if _, hasEntry := entries[goals[i].ID]; !hasEntry {
			return i
		}
	}
	
	// No incomplete goals found
	return currentIndex
}

// FindPreviousIncompleteGoal finds the previous goal that hasn't been entered yet.
func (n *NavigationHelper) FindPreviousIncompleteGoal(goals []models.Goal, entries map[string]models.GoalEntry, currentIndex int) int {
	// Start from the previous position
	for i := currentIndex - 1; i >= 0; i-- {
		if _, hasEntry := entries[goals[i].ID]; !hasEntry {
			return i
		}
	}
	
	// Wrap around to the end
	for i := len(goals) - 1; i >= currentIndex; i-- {
		if _, hasEntry := entries[goals[i].ID]; !hasEntry {
			return i
		}
	}
	
	// No incomplete goals found
	return currentIndex
}

// GetVisibleGoalsAfterFilter returns the list of goals that should be visible given the current filter state.
func (n *NavigationHelper) GetVisibleGoalsAfterFilter(goals []models.Goal, entries map[string]models.GoalEntry, filterState FilterState) []models.Goal {
	if filterState == FilterNone {
		return goals
	}
	
	var visibleGoals []models.Goal
	hideSkipped := filterState == FilterHideSkipped || filterState == FilterHideSkippedAndPrevious
	hidePrevious := filterState == FilterHidePrevious || filterState == FilterHideSkippedAndPrevious
	
	for _, goal := range goals {
		entry, hasEntry := entries[goal.ID]
		
		// Apply filter logic
		if hideSkipped && hasEntry && entry.Status == models.EntrySkipped {
			continue
		}
		
		if hidePrevious && hasEntry && (entry.Status == models.EntryCompleted || entry.Status == models.EntryFailed) {
			continue
		}
		
		visibleGoals = append(visibleGoals, goal)
	}
	
	return visibleGoals
}

// ShouldAutoSelectNext determines if we should automatically select the next incomplete goal.
func (n *NavigationHelper) ShouldAutoSelectNext(returnBehavior ReturnBehavior, justCompletedEntry bool) bool {
	return returnBehavior == ReturnToNextGoal && justCompletedEntry
}

// GetFilterDescription returns a human-readable description of the current filter state.
func (n *NavigationHelper) GetFilterDescription(filterState FilterState) string {
	switch filterState {
	case FilterNone:
		return "showing all goals"
	case FilterHideSkipped:
		return "hiding skipped goals"
	case FilterHidePrevious:
		return "hiding completed/failed goals"
	case FilterHideSkippedAndPrevious:
		return "hiding skipped and completed/failed goals"
	default:
		return "unknown filter"
	}
}

// Enhanced keybinding management with help text.

// GetShortHelp returns the short help bindings for the entry menu.
func (k *EntryMenuKeyMap) GetShortHelp() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.Select, k.NextIncomplete, k.ToggleReturnBehavior, k.FilterSkipped, k.FilterPrevious, k.Quit,
	}
}

// GetFullHelp returns the full help bindings for the entry menu.
func (k *EntryMenuKeyMap) GetFullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Navigation
		{k.Up, k.Down, k.Select},
		// Menu controls
		{k.ToggleReturnBehavior, k.FilterSkipped, k.FilterPrevious, k.ClearFilters},
		// Exit
		{k.Quit},
	}
}

// NavigationEnhancer provides additional navigation methods for the EntryMenuModel.
type NavigationEnhancer struct {
	helper *NavigationHelper
}

// NewNavigationEnhancer creates a new navigation enhancer.
func NewNavigationEnhancer() *NavigationEnhancer {
	return &NavigationEnhancer{
		helper: NewNavigationHelper(),
	}
}

// SelectNextIncompleteGoal selects the next incomplete goal in the list.
func (e *NavigationEnhancer) SelectNextIncompleteGoal(model *EntryMenuModel) {
	if len(model.goals) == 0 {
		return
	}
	
	currentIndex := model.list.Index()
	nextIndex := e.helper.FindNextIncompleteGoal(model.goals, model.entries, currentIndex)
	
	if nextIndex != currentIndex {
		model.list.Select(nextIndex)
	}
}

// SelectPreviousIncompleteGoal selects the previous incomplete goal in the list.
func (e *NavigationEnhancer) SelectPreviousIncompleteGoal(model *EntryMenuModel) {
	if len(model.goals) == 0 {
		return
	}
	
	currentIndex := model.list.Index()
	prevIndex := e.helper.FindPreviousIncompleteGoal(model.goals, model.entries, currentIndex)
	
	if prevIndex != currentIndex {
		model.list.Select(prevIndex)
	}
}

// UpdateListAfterFilterChange updates the list items and selection after a filter change.
func (e *NavigationEnhancer) UpdateListAfterFilterChange(model *EntryMenuModel) {
	// Get visible goals after filter
	visibleGoals := e.helper.GetVisibleGoalsAfterFilter(model.goals, model.entries, model.filterState)
	
	// Create menu items for visible goals
	var items []list.Item
	for _, goal := range visibleGoals {
		entry, hasEntry := model.entries[goal.ID]
		items = append(items, EntryMenuItem{
			Goal:             goal,
			EntryStatus:      entry.Status,
			HasEntry:         hasEntry,
			Value:            entry.Value,
			AchievementLevel: entry.AchievementLevel,
		})
	}
	
	// Update the list
	model.list.SetItems(items)
	
	// Auto-select first incomplete goal if list is not empty
	if len(items) > 0 {
		model.SelectFirstIncompleteGoal()
	}
}

// SelectFirstIncompleteGoal selects the first incomplete goal in the current list.
func (m *EntryMenuModel) SelectFirstIncompleteGoal() {
	items := m.list.Items()
	for i, item := range items {
		if menuItem, ok := item.(EntryMenuItem); ok {
			if !menuItem.HasEntry {
				m.list.Select(i)
				return
			}
		}
	}
	
	// If no incomplete goals, select first item
	if len(items) > 0 {
		m.list.Select(0)
	}
}

// GetCurrentGoalInfo returns information about the currently selected goal.
func (m *EntryMenuModel) GetCurrentGoalInfo() *GoalInfo {
	if len(m.goals) == 0 {
		return nil
	}
	
	selected := m.list.SelectedItem()
	if item, ok := selected.(EntryMenuItem); ok {
		entry, hasEntry := m.entries[item.Goal.ID]
		return &GoalInfo{
			Goal:     item.Goal,
			Entry:    entry,
			HasEntry: hasEntry,
			Index:    m.list.Index(),
		}
	}
	
	return nil
}

// GoalInfo contains information about a goal and its entry status.
type GoalInfo struct {
	Goal     models.Goal
	Entry    models.GoalEntry
	HasEntry bool
	Index    int
}

// IsComplete returns true if the goal has been completed.
func (g *GoalInfo) IsComplete() bool {
	return g.HasEntry && g.Entry.Status == models.EntryCompleted
}

// IsIncomplete returns true if the goal has not been entered yet.
func (g *GoalInfo) IsIncomplete() bool {
	return !g.HasEntry
}

// IsSkipped returns true if the goal has been skipped.
func (g *GoalInfo) IsSkipped() bool {
	return g.HasEntry && g.Entry.Status == models.EntrySkipped
}

// IsFailed returns true if the goal has failed.
func (g *GoalInfo) IsFailed() bool {
	return g.HasEntry && g.Entry.Status == models.EntryFailed
}