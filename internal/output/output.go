package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Truelist-io-Email-Validation/truelist-cli/internal/client"
	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen, color.Bold)
	red    = color.New(color.FgRed, color.Bold)
	yellow = color.New(color.FgYellow, color.Bold)
	cyan   = color.New(color.FgCyan)
	dim    = color.New(color.Faint)
	bold   = color.New(color.Bold)
)

// PrintValidationResult writes a human-readable validation result.
func PrintValidationResult(w io.Writer, r *client.ValidationResult) {
	icon, iconColor := stateIcon(r.State)
	iconColor.Fprintf(w, "%s %s\n", icon, r.Email)

	fmt.Fprintf(w, "  %-12s %s\n", dim.Sprint("State:"), stateColorized(r.State))
	fmt.Fprintf(w, "  %-12s %s\n", dim.Sprint("Sub-state:"), r.SubState)

	freeEmail := "no"
	if r.FreeEmail {
		freeEmail = "yes"
	}
	fmt.Fprintf(w, "  %-12s %s\n", dim.Sprint("Free email:"), freeEmail)

	if r.Suggestion != "" {
		fmt.Fprintf(w, "  %-12s %s\n", dim.Sprint("Suggestion:"), cyan.Sprint(r.Suggestion))
	}
}

// PrintValidationJSON writes the result as JSON.
func PrintValidationJSON(w io.Writer, r *client.ValidationResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// PrintValidationQuiet writes just the state string.
func PrintValidationQuiet(w io.Writer, r *client.ValidationResult) {
	fmt.Fprintln(w, r.State)
}

// PrintSummary writes a validation batch summary.
func PrintSummary(w io.Writer, total, valid, invalid, risky, unknown int) {
	fmt.Fprintln(w)
	bold.Fprintln(w, "Summary")
	fmt.Fprintf(w, "  Total:   %d\n", total)
	green.Fprintf(w, "  Valid:   %d\n", valid)
	red.Fprintf(w, "  Invalid: %d\n", invalid)
	yellow.Fprintf(w, "  Risky:   %d\n", risky)
	dim.Fprintf(w, "  Unknown: %d\n", unknown)
}

// PrintAccountInfo writes account details.
func PrintAccountInfo(w io.Writer, info *client.AccountInfo) {
	bold.Fprintln(w, "Account Info")
	fmt.Fprintf(w, "  %-10s %s\n", dim.Sprint("Email:"), info.Email)
	fmt.Fprintf(w, "  %-10s %s\n", dim.Sprint("Plan:"), info.Plan)
	fmt.Fprintf(w, "  %-10s %d\n", dim.Sprint("Credits:"), info.Credits)
}

// PrintError writes a user-friendly error message.
func PrintError(w io.Writer, msg string) {
	red.Fprintf(w, "Error: %s\n", msg)
}

func stateIcon(state string) (string, *color.Color) {
	switch strings.ToLower(state) {
	case "valid":
		return "\u2713", green
	case "invalid":
		return "\u2717", red
	case "risky":
		return "!", yellow
	default:
		return "?", dim
	}
}

func stateColorized(state string) string {
	switch strings.ToLower(state) {
	case "valid":
		return green.Sprint(state)
	case "invalid":
		return red.Sprint(state)
	case "risky":
		return yellow.Sprint(state)
	default:
		return dim.Sprint(state)
	}
}
