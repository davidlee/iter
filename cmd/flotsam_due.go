package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/davidlee/vice/internal/flotsam"
	"github.com/davidlee/vice/internal/srs"
)

var (
	// Due command flags
	dueFormat string // output format: table, json, paths
	dueLimit  int    // maximum number of results to show
)

// flotsamDueCmd represents the flotsam due command
// AIDEV-NOTE: T041/5.2-due-cmd; implements ZK-first enrichment pattern per ADR-008
// AIDEV-NOTE: shows due today + overdue notes, sorted by due date ascending then filename
var flotsamDueCmd = &cobra.Command{
	Use:   "due",
	Short: "Show flotsam notes due for review",
	Long: `Show flotsam notes that are due for spaced repetition review.

Displays notes that are due today or overdue, combining ZK metadata with SRS 
scheduling data. Follows the ZK-first enrichment pattern (ADR-008) for 
consistent data flow and rich metadata access.

Results are sorted by due date (oldest first), then by filename for deterministic
ordering when due dates are equal.

Examples:
  vice flotsam due                  # Show all due/overdue notes in table format
  vice flotsam due --format json   # Output in JSON format for scripting
  vice flotsam due --limit 10      # Show only first 10 results`,
	RunE: runFlotsamDue,
}

func init() {
	flotsamCmd.AddCommand(flotsamDueCmd)

	// Output format options
	flotsamDueCmd.Flags().StringVar(&dueFormat, "format", "table", "output format (table, json, paths)")
	flotsamDueCmd.Flags().IntVar(&dueLimit, "limit", 0, "maximum number of results (0 = no limit)")
}

// dueNote combines ZK metadata with SRS scheduling data for due notes
type dueNote struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Path     string    `json:"path"`
	DueDate  time.Time `json:"due_date"`
	Overdue  bool      `json:"overdue"`
	DaysPast int       `json:"days_past"`
}

// runFlotsamDue executes the flotsam due command using ZK-first enrichment pattern
func runFlotsamDue(_ *cobra.Command, _ []string) error {
	env := GetViceEnv()

	// Step 1: ZK query for note discovery and metadata (ZK-first pattern per ADR-008)
	notes, err := flotsam.GetAllViceNotes(env)
	if err != nil {
		return fmt.Errorf("failed to query vice-typed notes: %w", err)
	}

	if len(notes) == 0 {
		fmt.Println("No vice-typed notes found")
		return nil
	}

	// Step 2: Open SRS database for enrichment
	srsDB, err := srs.NewDatabase(env.ContextData, env.Context)
	if err != nil {
		return fmt.Errorf("failed to open SRS database: %w", err)
	}
	defer func() {
		if err := srsDB.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close SRS database: %v\n", err)
		}
	}()

	// Step 3: Enrich with SRS data and filter for due/overdue notes
	dueNotes, err := getDueNotes(notes, srsDB)
	if err != nil {
		return fmt.Errorf("failed to get due notes: %w", err)
	}

	// Step 4: Sort by due date (oldest first), then by filename for deterministic ordering
	sort.Slice(dueNotes, func(i, j int) bool {
		if dueNotes[i].DueDate.Equal(dueNotes[j].DueDate) {
			return dueNotes[i].Path < dueNotes[j].Path
		}
		return dueNotes[i].DueDate.Before(dueNotes[j].DueDate)
	})

	// Step 5: Apply limit if specified
	if dueLimit > 0 && len(dueNotes) > dueLimit {
		dueNotes = dueNotes[:dueLimit]
	}

	// Step 6: Format and output results
	return outputDueNotes(dueNotes, dueFormat)
}

// getDueNotes enriches note paths with SRS data and filters for due/overdue notes
func getDueNotes(notePaths []string, srsDB *srs.Database) ([]dueNote, error) {
	var dueNotes []dueNote
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	for _, notePath := range notePaths {
		// Get SRS scheduling data
		srsData, err := srsDB.GetSRSData(notePath)
		if err != nil {
			// Note exists in ZK but not in SRS database - skip silently
			// This handles notes that haven't been added to SRS yet
			continue
		}

		dueDate := time.Unix(srsData.Due, 0)

		// Filter: only include notes due today or overdue
		if dueDate.After(today) {
			continue // Note is due in the future
		}

		// Extract note ID and title from path
		noteID, title := extractNoteMetadata(notePath)

		// Calculate overdue status and days past
		overdue := dueDate.Before(today.Add(-24 * time.Hour)) // More than 1 day overdue
		daysPast := int(now.Sub(dueDate).Hours() / 24)
		if daysPast < 0 {
			daysPast = 0
		}

		dueNotes = append(dueNotes, dueNote{
			ID:       noteID,
			Title:    title,
			Path:     notePath,
			DueDate:  dueDate,
			Overdue:  overdue,
			DaysPast: daysPast,
		})
	}

	return dueNotes, nil
}

// extractNoteMetadata extracts note ID and title from file path
// This is a simple implementation - could be enhanced to parse ZK metadata
func extractNoteMetadata(notePath string) (string, string) {
	filename := filepath.Base(notePath)

	// Remove .md extension
	filename = strings.TrimSuffix(filename, ".md")

	// Try to extract ZK-style ID (first 4 chars if alphanumeric)
	if len(filename) >= 4 {
		possibleID := filename[:4]
		if isAlphanumeric(possibleID) {
			title := strings.TrimSpace(filename[4:])
			if strings.HasPrefix(title, "-") || strings.HasPrefix(title, "_") {
				title = strings.TrimSpace(title[1:])
			}
			if title == "" {
				title = filename // Fallback to full filename
			}
			return possibleID, title
		}
	}

	// Fallback: use filename as both ID and title
	return filename, filename
}

// isAlphanumeric checks if string contains only letters and numbers
func isAlphanumeric(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

// outputDueNotes formats and outputs due notes in the specified format
func outputDueNotes(notes []dueNote, format string) error {
	switch format {
	case "paths":
		for _, note := range notes {
			fmt.Println(note.Path)
		}
	case "table":
		if len(notes) == 0 {
			fmt.Println("No notes due for review")
			return nil
		}

		fmt.Printf("Found %d note(s) due for review:\n\n", len(notes))
		fmt.Printf("%-8s %-40s %-12s %-8s\n", "ID", "Title", "Due Date", "Status")
		fmt.Printf("%s\n", strings.Repeat("-", 70))

		for _, note := range notes {
			status := "Due today"
			if note.Overdue {
				if note.DaysPast == 1 {
					status = "1 day late"
				} else {
					status = fmt.Sprintf("%d days late", note.DaysPast)
				}
			}

			// Truncate title if too long
			title := note.Title
			if len(title) > 40 {
				title = title[:37] + "..."
			}

			fmt.Printf("%-8s %-40s %-12s %-8s\n",
				note.ID,
				title,
				note.DueDate.Format("2006-01-02"),
				status)
		}
	case "json":
		// Output as JSON array
		fmt.Print("[")
		for i, note := range notes {
			if i > 0 {
				fmt.Print(",")
			}
			fmt.Printf(`{"id":"%s","title":"%s","path":"%s","due_date":"%s","overdue":%t,"days_past":%d}`,
				note.ID, note.Title, note.Path, note.DueDate.Format(time.RFC3339), note.Overdue, note.DaysPast)
		}
		fmt.Println("]")
	default:
		return fmt.Errorf("invalid format: %s (valid: table, json, paths)", format)
	}
	return nil
}
