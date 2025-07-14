// Package main provides the vice CLI habit tracker application.
package main

import (
	"davidlee/vice/cmd"

	// Import dependencies to keep them in go.mod
	_ "github.com/charmbracelet/bubbles"
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/huh"
	_ "github.com/charmbracelet/lipgloss"
	_ "github.com/goccy/go-yaml"
	_ "github.com/stretchr/testify"
)

func main() {
	cmd.Execute()
}
