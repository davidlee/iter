package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/davidlee/vice/internal/flotsam"
	"github.com/davidlee/vice/internal/srs"
)

var (
	// Output format flags
	listFormat   string // output format: table, json, paths
	listTypeFlag string // filter by note type: flashcard, idea, script, log, all
	showSRS      bool   // include SRS scheduling information
)

// flotsamListCmd represents the flotsam list command
// AIDEV-NOTE: T041/5.1-list-cmd; combines zk delegation with SRS database for enriched note listing
// AIDEV-NOTE: implements type filtering, SRS enrichment, multiple output formats with graceful ZK degradation
var flotsamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List flotsam notes with optional SRS information",
	Long: `List flotsam notes by delegating to zk and enriching with SRS scheduling data.

By default, lists all notes with vice:type:* tags. Use --type to filter by specific
note types. Use --srs to include spaced repetition scheduling information.

The command delegates note discovery to zk and combines results with SRS database
queries for comprehensive note management information.

Examples:
  vice flotsam list                    # List all vice-typed notes
  vice flotsam list --type flashcard  # List only flashcard notes  
  vice flotsam list --srs             # Include SRS scheduling info
  vice flotsam list --format json     # Output in JSON format`,
	RunE: runFlotsamList,
}

func init() {
	flotsamCmd.AddCommand(flotsamListCmd)

	// Output format options
	flotsamListCmd.Flags().StringVar(&listFormat, "format", "table", "output format (table, json, paths)")
	flotsamListCmd.Flags().StringVar(&listTypeFlag, "type", "all", "note type filter (flashcard, idea, script, log, all)")
	flotsamListCmd.Flags().BoolVar(&showSRS, "srs", false, "include SRS scheduling information")
}

// runFlotsamList executes the flotsam list command
func runFlotsamList(_ *cobra.Command, _ []string) error {
	env := GetViceEnv()

	// Auto-initialize flotsam environment if needed
	if err := flotsam.EnsureFlotsamEnvironment(env); err != nil {
		return fmt.Errorf("failed to initialize flotsam environment: %w", err)
	}

	// Query notes based on type filter
	var notes []string
	var err error

	switch listTypeFlag {
	case "flashcard":
		notes, err = flotsam.GetFlashcardNotes(env)
	case "idea":
		notes, err = flotsam.GetIdeaNotes(env)
	case "script":
		notes, err = flotsam.GetScriptNotes(env)
	case "log":
		notes, err = flotsam.GetLogNotes(env)
	case "all":
		notes, err = flotsam.GetAllViceNotes(env)
	default:
		return fmt.Errorf("invalid note type: %s (valid: flashcard, idea, script, log, all)", listTypeFlag)
	}

	if err != nil {
		return fmt.Errorf("failed to query notes: %w", err)
	}

	// If no SRS info requested, output simple format
	if !showSRS {
		return outputNotes(notes, listFormat)
	}

	// Query SRS data for enriched output
	srsDB, err := srs.NewDatabase(env.ContextData, env.Context)
	if err != nil {
		return fmt.Errorf("failed to open SRS database: %w", err)
	}
	defer func() {
		if err := srsDB.Close(); err != nil {
			// Log error but don't fail the command
			fmt.Fprintf(os.Stderr, "Warning: failed to close SRS database: %v\n", err)
		}
	}()

	enrichedNotes, err := enrichNotesWithSRS(notes, srsDB)
	if err != nil {
		return fmt.Errorf("failed to enrich notes with SRS data: %w", err)
	}

	return outputEnrichedNotes(enrichedNotes, listFormat)
}

// outputNotes outputs notes in the specified format without SRS data
func outputNotes(notes []string, format string) error {
	switch format {
	case "paths":
		for _, note := range notes {
			fmt.Println(note)
		}
	case "table":
		if len(notes) == 0 {
			fmt.Println("No vice-typed notes found")
			return nil
		}
		fmt.Printf("Found %d note(s):\n\n", len(notes))
		for _, note := range notes {
			fmt.Printf("  %s\n", note)
		}
	case "json":
		// Simple JSON array of paths
		fmt.Print("[")
		for i, note := range notes {
			if i > 0 {
				fmt.Print(",")
			}
			fmt.Printf(`"%s"`, note)
		}
		fmt.Println("]")
	default:
		return fmt.Errorf("invalid format: %s (valid: table, json, paths)", format)
	}
	return nil
}

// enrichedNote combines note path with SRS scheduling data
type enrichedNote struct {
	Path               string     `json:"path"`
	HasSRS             bool       `json:"has_srs"`
	DueDate            *time.Time `json:"due_date,omitempty"`
	TotalReviews       int        `json:"total_reviews"`
	ConsecutiveCorrect int        `json:"consecutive_correct"`
	Easiness           float64    `json:"easiness"`
}

// enrichNotesWithSRS combines note paths with SRS scheduling data
func enrichNotesWithSRS(notes []string, srsDB *srs.Database) ([]enrichedNote, error) {
	enriched := make([]enrichedNote, 0, len(notes))

	for _, notePath := range notes {
		srsData, err := srsDB.GetSRSData(notePath)
		if err != nil {
			// Note exists but no SRS data - include with defaults
			enriched = append(enriched, enrichedNote{
				Path:   notePath,
				HasSRS: false,
			})
			continue
		}

		// Note has SRS data
		dueDate := time.Unix(srsData.Due, 0)
		enriched = append(enriched, enrichedNote{
			Path:               notePath,
			HasSRS:             true,
			DueDate:            &dueDate,
			TotalReviews:       srsData.TotalReviews,
			ConsecutiveCorrect: srsData.ConsecutiveCorrect,
			Easiness:           srsData.Easiness,
		})
	}

	return enriched, nil
}

// outputEnrichedNotes outputs enriched notes with SRS data in specified format
func outputEnrichedNotes(notes []enrichedNote, format string) error {
	switch format {
	case "paths":
		for _, note := range notes {
			fmt.Println(note.Path)
		}
	case "table":
		if len(notes) == 0 {
			fmt.Println("No vice-typed notes found")
			return nil
		}

		fmt.Printf("Found %d note(s) with SRS information:\n\n", len(notes))
		fmt.Printf("%-50s %-12s %-8s %-10s %-8s\n", "Path", "Next Due", "Reviews", "Correct", "Easiness")
		fmt.Printf("%s\n", strings.Repeat("-", 88))

		for _, note := range notes {
			if !note.HasSRS {
				fmt.Printf("%-50s %-12s %-8s %-10s %-8s\n", note.Path, "No SRS", "-", "-", "-")
				continue
			}

			dueStr := "Past due"
			if note.DueDate != nil {
				if note.DueDate.After(time.Now()) {
					dueStr = note.DueDate.Format("2006-01-02")
				}
			}

			fmt.Printf("%-50s %-12s %-8d %-10d %-8.1f\n",
				note.Path, dueStr, note.TotalReviews, note.ConsecutiveCorrect, note.Easiness)
		}
	case "json":
		// Output as JSON array
		fmt.Print("[")
		for i, note := range notes {
			if i > 0 {
				fmt.Print(",")
			}
			fmt.Printf(`{"path":"%s","has_srs":%t`, note.Path, note.HasSRS)
			if note.HasSRS && note.DueDate != nil {
				fmt.Printf(`,"due_date":"%s","total_reviews":%d,"consecutive_correct":%d,"easiness":%.1f`,
					note.DueDate.Format(time.RFC3339), note.TotalReviews, note.ConsecutiveCorrect, note.Easiness)
			}
			fmt.Print("}")
		}
		fmt.Println("]")
	default:
		return fmt.Errorf("invalid format: %s (valid: table, json, paths)", format)
	}
	return nil
}
