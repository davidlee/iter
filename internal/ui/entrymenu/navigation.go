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

// FindNextIncompleteHabit finds the next habit that hasn't been entered yet.
func (n *NavigationHelper) FindNextIncompleteHabit(habits []models.Habit, entries map[string]models.HabitEntry, currentIndex int) int {
	// Start from the next position after current
	for i := currentIndex + 1; i < len(habits); i++ {
		if _, hasEntry := entries[habits[i].ID]; !hasEntry {
			return i
		}
	}

	// Wrap around to the beginning
	for i := 0; i <= currentIndex; i++ {
		if _, hasEntry := entries[habits[i].ID]; !hasEntry {
			return i
		}
	}

	// No incomplete habits found
	return currentIndex
}

// FindPreviousIncompleteHabit finds the previous habit that hasn't been entered yet.
func (n *NavigationHelper) FindPreviousIncompleteHabit(habits []models.Habit, entries map[string]models.HabitEntry, currentIndex int) int {
	// Start from the previous position
	for i := currentIndex - 1; i >= 0; i-- {
		if _, hasEntry := entries[habits[i].ID]; !hasEntry {
			return i
		}
	}

	// Wrap around to the end
	for i := len(habits) - 1; i >= currentIndex; i-- {
		if _, hasEntry := entries[habits[i].ID]; !hasEntry {
			return i
		}
	}

	// No incomplete habits found
	return currentIndex
}

// GetVisibleHabitsAfterFilter returns the list of habits that should be visible given the current filter state.
func (n *NavigationHelper) GetVisibleHabitsAfterFilter(habits []models.Habit, entries map[string]models.HabitEntry, filterState FilterState) []models.Habit {
	if filterState == FilterNone {
		return habits
	}

	var visibleHabits []models.Habit
	hideSkipped := filterState == FilterHideSkipped || filterState == FilterHideSkippedAndPrevious
	hidePrevious := filterState == FilterHidePrevious || filterState == FilterHideSkippedAndPrevious

	for _, habit := range habits {
		entry, hasEntry := entries[habit.ID]

		// Apply filter logic
		if hideSkipped && hasEntry && entry.Status == models.EntrySkipped {
			continue
		}

		if hidePrevious && hasEntry && (entry.Status == models.EntryCompleted || entry.Status == models.EntryFailed) {
			continue
		}

		visibleHabits = append(visibleHabits, habit)
	}

	return visibleHabits
}

// ShouldAutoSelectNext determines if we should automatically select the next incomplete habit.
func (n *NavigationHelper) ShouldAutoSelectNext(returnBehavior ReturnBehavior, justCompletedEntry bool) bool {
	return returnBehavior == ReturnToNextHabit && justCompletedEntry
}

// GetFilterDescription returns a human-readable description of the current filter state.
func (n *NavigationHelper) GetFilterDescription(filterState FilterState) string {
	switch filterState {
	case FilterNone:
		return "showing all habits"
	case FilterHideSkipped:
		return "hiding skipped habits"
	case FilterHidePrevious:
		return "hiding completed/failed habits"
	case FilterHideSkippedAndPrevious:
		return "hiding skipped and completed/failed habits"
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

// SelectNextIncompleteHabit selects the next incomplete habit in the list.
func (e *NavigationEnhancer) SelectNextIncompleteHabit(model *EntryMenuModel) {
	if len(model.habits) == 0 {
		return
	}

	currentIndex := model.list.Index()
	nextIndex := e.helper.FindNextIncompleteHabit(model.habits, model.entries, currentIndex)

	if nextIndex != currentIndex {
		model.list.Select(nextIndex)
	}
}

// SelectPreviousIncompleteHabit selects the previous incomplete habit in the list.
func (e *NavigationEnhancer) SelectPreviousIncompleteHabit(model *EntryMenuModel) {
	if len(model.habits) == 0 {
		return
	}

	currentIndex := model.list.Index()
	prevIndex := e.helper.FindPreviousIncompleteHabit(model.habits, model.entries, currentIndex)

	if prevIndex != currentIndex {
		model.list.Select(prevIndex)
	}
}

// UpdateListAfterFilterChange updates the list items and selection after a filter change.
func (e *NavigationEnhancer) UpdateListAfterFilterChange(model *EntryMenuModel) {
	// Get visible habits after filter
	visibleHabits := e.helper.GetVisibleHabitsAfterFilter(model.habits, model.entries, model.filterState)

	// Create menu items for visible habits
	var items []list.Item
	for _, habit := range visibleHabits {
		entry, hasEntry := model.entries[habit.ID]
		items = append(items, EntryMenuItem{
			Habit:            habit,
			EntryStatus:      entry.Status,
			HasEntry:         hasEntry,
			Value:            entry.Value,
			AchievementLevel: entry.AchievementLevel,
		})
	}

	// Update the list
	model.list.SetItems(items)

	// Auto-select first incomplete habit if list is not empty
	if len(items) > 0 {
		model.SelectFirstIncompleteHabit()
	}
}

// SelectFirstIncompleteHabit selects the first incomplete habit in the current list.
func (m *EntryMenuModel) SelectFirstIncompleteHabit() {
	items := m.list.Items()
	for i, item := range items {
		if menuItem, ok := item.(EntryMenuItem); ok {
			if !menuItem.HasEntry {
				m.list.Select(i)
				return
			}
		}
	}

	// If no incomplete habits, select first item
	if len(items) > 0 {
		m.list.Select(0)
	}
}

// GetCurrentHabitInfo returns information about the currently selected habit.
func (m *EntryMenuModel) GetCurrentHabitInfo() *HabitInfo {
	if len(m.habits) == 0 {
		return nil
	}

	selected := m.list.SelectedItem()
	if item, ok := selected.(EntryMenuItem); ok {
		entry, hasEntry := m.entries[item.Habit.ID]
		return &HabitInfo{
			Habit:    item.Habit,
			Entry:    entry,
			HasEntry: hasEntry,
			Index:    m.list.Index(),
		}
	}

	return nil
}

// HabitInfo contains information about a habit and its entry status.
type HabitInfo struct {
	Habit    models.Habit
	Entry    models.HabitEntry
	HasEntry bool
	Index    int
}

// IsComplete returns true if the habit has been completed.
func (g *HabitInfo) IsComplete() bool {
	return g.HasEntry && g.Entry.Status == models.EntryCompleted
}

// IsIncomplete returns true if the habit has not been entered yet.
func (g *HabitInfo) IsIncomplete() bool {
	return !g.HasEntry
}

// IsSkipped returns true if the habit has been skipped.
func (g *HabitInfo) IsSkipped() bool {
	return g.HasEntry && g.Entry.Status == models.EntrySkipped
}

// IsFailed returns true if the habit has failed.
func (g *HabitInfo) IsFailed() bool {
	return g.HasEntry && g.Entry.Status == models.EntryFailed
}
