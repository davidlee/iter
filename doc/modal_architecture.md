# Modal System Architecture for Entry Forms

## Overview

This document outlines the design for a modal/overlay system in the vice application to replace the current `form.Run()` takeover approach for entry forms. The modal system will eliminate the edit looping bug and provide a better user experience.

<!-- AIDEV-NOTE: modal-architecture; comprehensive design document for T024 bug fixes -->
<!-- AIDEV-NOTE: T024-solution-design; architectural approach eliminates form.Run() complexity -->

## Research Findings

### BubbleTea Modal Patterns

From analyzing the charmbracelet ecosystem:

1. **No built-in modal system**: BubbleTea doesn't have native modal support
2. **Composable views pattern**: `examples/composable-views` shows state-based view switching
3. **Multi-view pattern**: `examples/views` demonstrates view transitions with state management
4. **Viewport component**: `bubbles/viewport` provides scrollable content areas
5. **Huh forms**: Use `form.Run()` for complete takeover (current problematic approach)

### Current Problem Analysis

**Current Flow** (problematic):
```
EntryMenu → CollectSingleGoalEntry() → flow.CollectEntry() → form.Run() → [TAKEOVER]
```

**Desired Flow** (modal):
```
EntryMenu + Modal(EntryForm) → Close Modal → EntryMenu (updated)
```

## Modal Architecture Design

### Core Components

#### 1. Modal Interface
```go
type Modal interface {
    // Lifecycle
    Init() tea.Cmd
    Update(msg tea.Msg) (Modal, tea.Cmd)
    View() string
    
    // State
    IsOpen() bool
    IsClosed() bool
    
    // Focus management
    HandleKey(msg tea.KeyMsg) (Modal, tea.Cmd)
    
    // Integration
    GetResult() interface{}
}
```

#### 2. Modal Manager
```go
type ModalManager struct {
    activeModal Modal
    parentModel tea.Model
    backgroundView string
    overlayStyle lipgloss.Style
}
```

#### 3. Entry Form Modal
```go
type EntryFormModal struct {
    goal models.Goal
    collector *ui.EntryCollector
    form *huh.Form
    result *EntryResult
    state ModalState
}
```

### State Management

#### Modal States
```go
type ModalState int

const (
    ModalOpening ModalState = iota
    ModalActive
    ModalClosing
    ModalClosed
)
```

#### Integration Pattern
```go
// Parent model (EntryMenuModel) integrates modal
type EntryMenuModel struct {
    // ... existing fields
    modalManager *ModalManager
    showModal    bool
}

// Modal events
type ModalOpenedMsg struct{ Modal Modal }
type ModalClosedMsg struct{ Result interface{} }
```

### Rendering Strategy

#### Layered Rendering
1. **Background**: Original menu view (dimmed/blurred)
2. **Overlay**: Modal content with border/shadow
3. **Focus**: Modal captures all keyboard input

#### Layout Approach
```go
func (m *EntryMenuModel) View() string {
    baseView := m.renderMenuView()
    
    if m.showModal && m.modalManager.HasActiveModal() {
        modalView := m.modalManager.View()
        return m.renderWithModal(baseView, modalView)
    }
    
    return baseView
}

func (m *EntryMenuModel) renderWithModal(background, modal string) string {
    // Dim background
    dimmedBg := m.dimStyle.Render(background)
    
    // Center modal
    centeredModal := lipgloss.Place(
        m.width, m.height,
        lipgloss.Center, lipgloss.Center,
        m.modalStyle.Render(modal),
    )
    
    // Overlay modal on background
    return lipgloss.JoinVertical(lipgloss.Left, dimmedBg, centeredModal)
}
```

### Event Flow

#### Opening Modal
1. User presses Enter on goal
2. `EntryMenuModel.Update()` receives key
3. Create `EntryFormModal` with goal
4. `ModalManager.OpenModal(modal)`
5. Modal renders on top of menu

#### Modal Interaction
1. All keys routed to modal
2. Modal updates internal form state
3. ESC or completion closes modal
4. Modal returns result

#### Closing Modal
1. Modal sends `ModalClosedMsg` with result
2. `EntryMenuModel` receives message
3. Updates entries from result
4. Auto-saves if configured
5. Returns to normal menu view

### Key Benefits

<!-- AIDEV-NOTE: modal-benefits; architectural advantages over form.Run() approach -->
1. **Eliminates Handoff**: No complex state transfer between menu and form
2. **Natural UX**: Modal close → return to menu (no looping)
3. **Context Preservation**: User sees menu behind modal
4. **Clean Architecture**: Clear separation of concerns
5. **Extensible**: Modal system can be reused for other dialogs

## Implementation Strategy

### Phase 1: Core Modal Infrastructure
- `internal/ui/modal/` package
- Basic modal interface and manager
- Simple overlay rendering

### Phase 2: Entry Form Modal
- `EntryFormModal` implementation
- Integration with existing `EntryFieldInput` components
- Replace `form.Run()` with modal approach

### Phase 3: Menu Integration
- Modify `EntryMenuModel` for modal support
- Event routing and state management
- Rendering with overlay

### Phase 4: Testing & Refinement
- teatest integration tests
- Modal interaction testing
- Bug verification and cleanup

## Technical Considerations

### Keyboard Navigation
- Modal captures all keyboard input when active
- ESC key closes modal (standardized)
- Tab/Enter navigation within modal
- Background menu stays visible but inactive

### Performance
- Minimal impact: modal is lightweight overlay
- No need for complex state synchronization
- Background view rendered once, modal overlaid

### Error Handling
- Modal errors displayed within modal
- Graceful fallback to menu on modal failure
- Proper cleanup on unexpected modal close

### Testing Strategy
- Unit tests for modal components
- Integration tests for modal lifecycle
- teatest for complete user flows
- Regression tests for original bugs

## Migration Plan

### Current Dependencies
- `form.Run()` calls in goal collection flows
- Direct form interaction in input components
- State synchronization logic in entry menu

### Migration Steps
1. Implement modal system alongside current system
2. Create modal versions of entry forms
3. Update goal collection flows to use modals
4. Remove old `form.Run()` approach
5. Clean up unused synchronization code

This modal architecture will provide a superior user experience while eliminating the current bugs and simplifying the codebase.