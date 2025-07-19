package cmd

import (
	"github.com/spf13/cobra"
)

// flotsamCmd represents the flotsam command
// AIDEV-NOTE: T041/5.1-flotsam-cmd; Unix interop foundation for flotsam operations via zk delegation
var flotsamCmd = &cobra.Command{
	Use:   "flotsam",
	Short: "Manage flotsam notes with SRS scheduling",
	Long: `Manage flotsam (Zettelkasten + SRS) notes through Unix tool delegation.
Flotsam combines markdown note-taking with spaced repetition scheduling,
delegating core operations to zk while maintaining SRS state in vice.

Vice manages SRS scheduling data while zk handles note content, links, and search.
All notes with vice:type:* tags participate in spaced repetition learning.

Examples:
  vice flotsam list     # List all vice-typed notes with SRS status
  vice flotsam due      # Show notes due for review
  vice flotsam edit     # Edit notes via zk integration`,
}

func init() {
	rootCmd.AddCommand(flotsamCmd)
}
