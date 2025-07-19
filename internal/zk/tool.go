package zk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ToolErrorType represents different categories of tool execution errors
type ToolErrorType string

const (
	ErrToolNotFound      ToolErrorType = "tool_not_found"
	ErrToolNotExecutable ToolErrorType = "tool_not_executable"
	ErrCommandFailed     ToolErrorType = "command_failed"
	ErrInvalidOutput     ToolErrorType = "invalid_output"
	ErrPermissionDenied  ToolErrorType = "permission_denied"
)

// ToolError represents an error that occurred during tool execution
type ToolError struct {
	Type     ToolErrorType
	Tool     string
	Command  []string
	Message  string
	Cause    error
	ExitCode int
}

func (e *ToolError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("%s command failed: %s", e.Tool, strings.Join(e.Command, " "))
}

// NewToolError creates a new ToolError
func NewToolError(errorType ToolErrorType, tool string, command []string, cause error, exitCode int) *ToolError {
	return &ToolError{
		Type:     errorType,
		Tool:     tool,
		Command:  command,
		Cause:    cause,
		ExitCode: exitCode,
	}
}

// ToolResult encapsulates command execution results
type ToolResult struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Duration time.Duration
}

// Tool represents an external CLI tool for delegation
type Tool interface {
	Name() string
	IsAvailable() bool
	Execute(ctx context.Context, args ...string) (*ToolResult, error)
	Version() (string, error)
}

// ZKTool implements Tool interface for zk operations
type ZKTool struct {
	path        string
	workingDir  string
	notebookDir string
	env         []string
	globalFlags []string
}

// NewZKTool creates a new ZKTool instance
func NewZKTool() *ZKTool {
	tool := &ZKTool{
		env: os.Environ(),
	}

	// Find zk binary in PATH
	if path, err := exec.LookPath("zk"); err == nil {
		tool.path = path
	}

	return tool
}

// NewZKToolWithPath creates a ZKTool with a specific binary path
func NewZKToolWithPath(zkPath string) *ZKTool {
	return &ZKTool{
		path: zkPath,
		env:  os.Environ(),
	}
}

// SetWorkingDir sets the working directory for zk commands
func (zk *ZKTool) SetWorkingDir(dir string) {
	zk.workingDir = dir
	zk.updateGlobalFlags()
}

// SetNotebookDir sets the notebook directory via --notebook-dir flag
func (zk *ZKTool) SetNotebookDir(dir string) {
	zk.notebookDir = dir
	zk.updateGlobalFlags()
}

// updateGlobalFlags rebuilds the global flags array
func (zk *ZKTool) updateGlobalFlags() {
	zk.globalFlags = []string{}

	if zk.notebookDir != "" {
		zk.globalFlags = append(zk.globalFlags, "--notebook-dir", zk.notebookDir)
	}

	if zk.workingDir != "" {
		zk.globalFlags = append(zk.globalFlags, "--working-dir", zk.workingDir)
	}
}

// Name returns the tool name
func (zk *ZKTool) Name() string {
	return "zk"
}

// IsAvailable checks if zk is available for execution
func (zk *ZKTool) IsAvailable() bool {
	if zk.path == "" {
		return false
	}

	// Also check if the file exists and is executable
	if info, err := os.Stat(zk.path); err != nil || info.IsDir() {
		return false
	}

	return true
}

// Version returns the zk version
func (zk *ZKTool) Version() (string, error) {
	if !zk.IsAvailable() {
		return "", NewToolError(ErrToolNotFound, "zk", []string{"--version"}, nil, 0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := zk.Execute(ctx, "--version")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(result.Stdout)), nil
}

// Execute runs a zk command with the given arguments
func (zk *ZKTool) Execute(ctx context.Context, args ...string) (*ToolResult, error) {
	if !zk.IsAvailable() {
		return nil, NewToolError(ErrToolNotFound, "zk", args,
			fmt.Errorf("zk command not found in PATH"), 0)
	}

	// Sanitize arguments to prevent command injection
	sanitizedArgs := make([]string, len(args))
	for i, arg := range args {
		sanitizedArgs[i] = sanitizeArg(arg)
	}

	cmd := exec.CommandContext(ctx, zk.path, sanitizedArgs...)

	if zk.workingDir != "" {
		cmd.Dir = zk.workingDir
	}
	cmd.Env = zk.env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &ToolResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: 0,
		Duration: duration,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}

		return result, NewToolError(ErrCommandFailed, "zk", sanitizedArgs, err, result.ExitCode)
	}

	return result, nil
}

// sanitizeArg removes potentially dangerous characters from arguments
// This is a defense-in-depth measure since we use exec.Command (not shell)
func sanitizeArg(arg string) string {
	// Remove shell metacharacters as extra precaution
	dangerous := []string{";", "&", "|", "`", "$", "(", ")", "<", ">"}
	for _, char := range dangerous {
		arg = strings.ReplaceAll(arg, char, "")
	}
	return arg
}

// ZKCommand provides a fluent interface for building zk commands
type ZKCommand struct {
	tool        *ZKTool
	command     string
	args        []string
	notebookDir string
	workingDir  string
	noInput     bool
	format      string
	interactive bool
	limit       int
	tags        []string
	match       []string
	paths       []string
}

// List creates a new list command builder
func (zk *ZKTool) List() *ZKCommand {
	return &ZKCommand{
		tool:    zk,
		command: "list",
		args:    []string{"list"},
	}
}

// Edit creates a new edit command builder
func (zk *ZKTool) Edit() *ZKCommand {
	return &ZKCommand{
		tool:    zk,
		command: "edit",
		args:    []string{"edit"},
	}
}

// New creates a new note creation command builder
func (zk *ZKTool) New() *ZKCommand {
	return &ZKCommand{
		tool:    zk,
		command: "new",
		args:    []string{"new"},
	}
}

// Tag creates a new tag command builder
func (zk *ZKTool) Tag() *ZKCommand {
	return &ZKCommand{
		tool:    zk,
		command: "tag",
		args:    []string{"tag"},
	}
}

// Index creates a new index command builder
func (zk *ZKTool) Index() *ZKCommand {
	return &ZKCommand{
		tool:    zk,
		command: "index",
		args:    []string{"index"},
	}
}

// Command builder methods

// NotebookDir sets the --notebook-dir flag
func (cmd *ZKCommand) NotebookDir(dir string) *ZKCommand {
	cmd.notebookDir = dir
	return cmd
}

// WorkingDir sets the --working-dir flag
func (cmd *ZKCommand) WorkingDir(dir string) *ZKCommand {
	cmd.workingDir = dir
	return cmd
}

// NoInput sets the --no-input flag
func (cmd *ZKCommand) NoInput() *ZKCommand {
	cmd.noInput = true
	return cmd
}

// Format sets the --format flag (for list command)
func (cmd *ZKCommand) Format(format string) *ZKCommand {
	cmd.format = format
	return cmd
}

// Interactive sets the --interactive flag
func (cmd *ZKCommand) Interactive() *ZKCommand {
	cmd.interactive = true
	return cmd
}

// Limit sets the --limit flag
func (cmd *ZKCommand) Limit(count int) *ZKCommand {
	cmd.limit = count
	return cmd
}

// Tag adds tag filters (for list command)
func (cmd *ZKCommand) Tag(tags ...string) *ZKCommand {
	cmd.tags = append(cmd.tags, tags...)
	return cmd
}

// Match adds match filters (for list command)
func (cmd *ZKCommand) Match(queries ...string) *ZKCommand {
	cmd.match = append(cmd.match, queries...)
	return cmd
}

// Paths adds path arguments
func (cmd *ZKCommand) Paths(paths ...string) *ZKCommand {
	cmd.paths = append(cmd.paths, paths...)
	return cmd
}

// CreatedAfter adds --created-after filter
func (cmd *ZKCommand) CreatedAfter(date string) *ZKCommand {
	cmd.args = append(cmd.args, "--created-after", date)
	return cmd
}

// CreatedBefore adds --created-before filter
func (cmd *ZKCommand) CreatedBefore(date string) *ZKCommand {
	cmd.args = append(cmd.args, "--created-before", date)
	return cmd
}

// LinkedBy adds --linked-by filter
func (cmd *ZKCommand) LinkedBy(paths ...string) *ZKCommand {
	for _, path := range paths {
		cmd.args = append(cmd.args, "--linked-by", path)
	}
	return cmd
}

// LinkTo adds --link-to filter
func (cmd *ZKCommand) LinkTo(paths ...string) *ZKCommand {
	for _, path := range paths {
		cmd.args = append(cmd.args, "--link-to", path)
	}
	return cmd
}

// Title adds --title flag (for new command)
func (cmd *ZKCommand) Title(title string) *ZKCommand {
	cmd.args = append(cmd.args, "--title", title)
	return cmd
}

// Template adds --template flag (for new command)
func (cmd *ZKCommand) Template(template string) *ZKCommand {
	cmd.args = append(cmd.args, "--template", template)
	return cmd
}

// Extra adds --extra flag (for new command)
func (cmd *ZKCommand) Extra(extras ...string) *ZKCommand {
	for _, extra := range extras {
		cmd.args = append(cmd.args, "--extra", extra)
	}
	return cmd
}

// PrintPath adds --print-path flag (for new command)
func (cmd *ZKCommand) PrintPath() *ZKCommand {
	cmd.args = append(cmd.args, "--print-path")
	return cmd
}

// DryRun adds --dry-run flag (for new command)
func (cmd *ZKCommand) DryRun() *ZKCommand {
	cmd.args = append(cmd.args, "--dry-run")
	return cmd
}

// Execute builds and executes the command
func (cmd *ZKCommand) Execute(ctx context.Context) (*ToolResult, error) {
	// Start with tool's global flags (notebook-dir, working-dir from tool instance)
	args := make([]string, 0, len(cmd.tool.globalFlags)+len(cmd.args)+10)
	args = append(args, cmd.args...)
	args = append(args, cmd.tool.globalFlags...)

	// Add command-specific global flags (these override tool instance flags)
	if cmd.notebookDir != "" {
		args = append(args, "--notebook-dir", cmd.notebookDir)
	}
	if cmd.workingDir != "" {
		args = append(args, "--working-dir", cmd.workingDir)
	}
	if cmd.noInput {
		args = append(args, "--no-input")
	}

	// Add command-specific flags
	if cmd.format != "" {
		args = append(args, "--format", cmd.format)
	}
	if cmd.interactive {
		args = append(args, "--interactive")
	}
	if cmd.limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", cmd.limit))
	}

	// Add filters
	for _, tag := range cmd.tags {
		args = append(args, "--tag", tag)
	}
	for _, match := range cmd.match {
		args = append(args, "--match", match)
	}

	// Add paths at the end
	args = append(args, cmd.paths...)

	return cmd.tool.Execute(ctx, args...)
}

// Helper functions for parsing zk output

// ParseZKListJSON parses zk list output in JSON format
func ParseZKListJSON(data []byte) ([]ZKNote, error) {
	var notes []ZKNote
	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, fmt.Errorf("failed to parse zk JSON output: %w", err)
	}
	return notes, nil
}

// ParseZKPaths parses zk output in path format
func ParseZKPaths(data []byte) ([]string, error) {
	content := strings.TrimSpace(string(data))
	if content == "" {
		return []string{}, nil
	}
	return strings.Split(content, "\n"), nil
}

// ZKNote represents a note as returned by zk list --format json
type ZKNote struct {
	Filename   string                 `json:"filename"`
	FilePath   string                 `json:"file-path"`
	Title      string                 `json:"title"`
	Lead       string                 `json:"lead"`
	Body       string                 `json:"body"`
	RawContent string                 `json:"raw-content"`
	WordCount  int                    `json:"word-count"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
	Created    time.Time              `json:"created"`
	Modified   time.Time              `json:"modified"`
}

// HasTag checks if the note has a specific tag
func (n *ZKNote) HasTag(tag string) bool {
	for _, t := range n.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// HasSRS checks if the note participates in SRS
func (n *ZKNote) HasSRS() bool {
	return n.HasTag("vice:srs")
}

// HasType checks if the note has a specific vice type
func (n *ZKNote) HasType(noteType string) bool {
	return n.HasTag("vice:type:" + noteType)
}

// IsFlashcard checks if the note is a flashcard
func (n *ZKNote) IsFlashcard() bool {
	return n.HasType("flashcard")
}

// ValidateNotePath validates that a path is within the notebook directory
func ValidateNotePath(notebookDir, notePath string) error {
	absNotebook, err := filepath.Abs(notebookDir)
	if err != nil {
		return fmt.Errorf("failed to resolve notebook directory: %w", err)
	}

	absNote, err := filepath.Abs(notePath)
	if err != nil {
		return fmt.Errorf("failed to resolve note path: %w", err)
	}

	if !strings.HasPrefix(absNote, absNotebook) {
		return fmt.Errorf("path %s is outside notebook directory %s", notePath, notebookDir)
	}

	return nil
}
