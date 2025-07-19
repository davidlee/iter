package zk

import (
	"fmt"
	"time"
)

// MockZKTool provides a mock implementation of ZKTool for testing.
// AIDEV-NOTE: T041/6.1-testing-mock; provides controllable ZK behavior for unit tests
type MockZKTool struct {
	IsAvailableFlag bool
	InitCalled      bool
	InitReturns     MockExecuteResult

	ListCalled  bool
	ListReturns MockListResult

	EditCalled  bool
	EditReturns error

	GetLinkedNotesCalled  bool
	GetLinkedNotesReturns MockLinkedNotesResult

	ExecuteCalled  bool
	ExecuteReturns MockExecuteResult
}

// MockExecuteResult defines the return values for Execute method.
type MockExecuteResult struct {
	Output string
	Error  error
}

// MockListResult defines the return values for List method.
type MockListResult struct {
	Notes []string
	Error error
}

// MockLinkedNotesResult defines the return values for GetLinkedNotes method.
type MockLinkedNotesResult struct {
	Backlinks []string
	Outbound  []string
	Error     error
}

// Name returns the tool name.
func (m *MockZKTool) Name() string {
	return "zk"
}

// Available returns the configured availability status.
func (m *MockZKTool) Available() bool {
	return m.IsAvailableFlag
}

// Execute records the call and returns configured result.
func (m *MockZKTool) Execute(args ...string) (*ToolResult, error) {
	m.ExecuteCalled = true

	// Special handling for init command
	if len(args) > 0 && args[0] == "init" {
		m.InitCalled = true
		if m.InitReturns.Error != nil {
			return nil, m.InitReturns.Error
		}
		return &ToolResult{
			Stdout:   m.InitReturns.Output,
			Stderr:   "",
			ExitCode: 0,
			Duration: time.Millisecond,
		}, nil
	}

	if m.ExecuteReturns.Error != nil {
		return nil, m.ExecuteReturns.Error
	}

	return &ToolResult{
		Stdout:   m.ExecuteReturns.Output,
		Stderr:   "",
		ExitCode: 0,
		Duration: time.Millisecond,
	}, nil
}

// List records the call and returns configured result.
func (m *MockZKTool) List(_ ...string) ([]string, error) {
	m.ListCalled = true
	return m.ListReturns.Notes, m.ListReturns.Error
}

// Edit records the call and returns configured result.
func (m *MockZKTool) Edit(_ ...string) error {
	m.EditCalled = true
	return m.EditReturns
}

// GetLinkedNotes records the call and returns configured result.
func (m *MockZKTool) GetLinkedNotes(_ string) ([]string, []string, error) {
	m.GetLinkedNotesCalled = true
	return m.GetLinkedNotesReturns.Backlinks,
		m.GetLinkedNotesReturns.Outbound,
		m.GetLinkedNotesReturns.Error
}

// NewMockZKTool creates a new mock ZK tool with default configuration.
func NewMockZKTool() *MockZKTool {
	return &MockZKTool{
		IsAvailableFlag: true,
		InitReturns: MockExecuteResult{
			Output: "ZK notebook initialized",
			Error:  nil,
		},
		ListReturns: MockListResult{
			Notes: []string{},
			Error: nil,
		},
		EditReturns: nil,
		GetLinkedNotesReturns: MockLinkedNotesResult{
			Backlinks: []string{},
			Outbound:  []string{},
			Error:     nil,
		},
		ExecuteReturns: MockExecuteResult{
			Output: "command executed",
			Error:  nil,
		},
	}
}

// NewUnavailableMockZKTool creates a mock ZK tool that simulates ZK being unavailable.
func NewUnavailableMockZKTool() *MockZKTool {
	mock := NewMockZKTool()
	mock.IsAvailableFlag = false
	mock.ExecuteReturns = MockExecuteResult{
		Output: "",
		Error:  fmt.Errorf("zk command not found"),
	}
	return mock
}
