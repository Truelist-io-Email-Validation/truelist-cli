package cmd

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Truelist-io-Email-Validation/truelist-cli/internal/client"
	"github.com/Truelist-io-Email-Validation/truelist-cli/internal/config"
	"github.com/Truelist-io-Email-Validation/truelist-cli/internal/output"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	flagFile       string
	flagOutput     string
	flagColumn     string
	flagJSON       bool
	flagQuiet      bool
)

func init() {
	validateCmd.Flags().StringVarP(&flagFile, "file", "f", "", "CSV file of emails to validate")
	validateCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Output file path (default: <input>_validated.csv)")
	validateCmd.Flags().StringVarP(&flagColumn, "column", "c", "", "Name of the email column in the CSV")
	validateCmd.Flags().BoolVar(&flagJSON, "json", false, "Output results as JSON")
	validateCmd.Flags().BoolVarP(&flagQuiet, "quiet", "q", false, "Output only the state (valid/invalid/risky/unknown)")

	rootCmd.AddCommand(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate [email]",
	Short: "Validate one or more email addresses",
	Long: `Validate email addresses using the Truelist API.

Single email:
  truelist validate user@example.com

CSV file:
  truelist validate --file emails.csv

Stdin (pipe):
  cat emails.txt | truelist validate`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, err := config.GetAPIKey()
		if err != nil {
			output.PrintError(os.Stderr, err.Error())
			return err
		}

		c := client.New(apiKey)

		// Determine mode: file, stdin, or single email.
		switch {
		case flagFile != "":
			return runFileValidation(c)
		case len(args) == 0:
			return runStdinValidation(c)
		default:
			return runSingleValidation(c, args[0])
		}
	},
}

func runSingleValidation(c *client.Client, email string) error {
	result, err := c.Validate(context.Background(), email)
	if err != nil {
		output.PrintError(os.Stderr, err.Error())
		return err
	}

	switch {
	case flagJSON:
		return output.PrintValidationJSON(os.Stdout, result)
	case flagQuiet:
		output.PrintValidationQuiet(os.Stdout, result)
	default:
		output.PrintValidationResult(os.Stdout, result)
	}
	return nil
}

func runStdinValidation(c *client.Client) error {
	// Check if stdin is a pipe.
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf("no email provided — pass an email as an argument, use --file, or pipe from stdin")
	}

	scanner := bufio.NewScanner(os.Stdin)
	var results []*client.ValidationResult
	counts := map[string]int{"valid": 0, "invalid": 0, "risky": 0, "unknown": 0}

	for scanner.Scan() {
		email := strings.TrimSpace(scanner.Text())
		if email == "" {
			continue
		}

		result, err := c.Validate(context.Background(), email)
		if err != nil {
			output.PrintError(os.Stderr, fmt.Sprintf("failed to validate %s: %s", email, err))
			continue
		}

		results = append(results, result)
		counts[strings.ToLower(result.State)]++

		if flagJSON {
			// In JSON mode, we'll collect and print at the end.
			continue
		}
		if flagQuiet {
			output.PrintValidationQuiet(os.Stdout, result)
		} else {
			output.PrintValidationResult(os.Stdout, result)
			fmt.Println()
		}
	}

	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(results); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	}

	if !flagQuiet && !flagJSON {
		total := counts["valid"] + counts["invalid"] + counts["risky"] + counts["unknown"]
		output.PrintSummary(os.Stdout, total, counts["valid"], counts["invalid"], counts["risky"], counts["unknown"])
	}

	return scanner.Err()
}

func runFileValidation(c *client.Client) error {
	if flagJSON {
		return fmt.Errorf("--json flag is not supported with --file mode (CSV output is always used)")
	}
	if flagQuiet {
		return fmt.Errorf("--quiet flag is not supported with --file mode (CSV output is always used)")
	}

	f, err := os.Open(flagFile)
	if err != nil {
		output.PrintError(os.Stderr, fmt.Sprintf("could not open file: %s", err))
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// Read header row.
	header, err := reader.Read()
	if err != nil {
		output.PrintError(os.Stderr, fmt.Sprintf("could not read CSV header: %s", err))
		return err
	}

	emailColIdx := findEmailColumn(header, flagColumn)
	if emailColIdx == -1 {
		errMsg := "could not detect email column — use --column to specify it"
		if flagColumn != "" {
			errMsg = fmt.Sprintf("column %q not found in CSV header", flagColumn)
		}
		output.PrintError(os.Stderr, errMsg)
		return fmt.Errorf(errMsg)
	}

	// Read all rows so we can show a progress bar.
	var rows [][]string
	for {
		row, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			output.PrintError(os.Stderr, fmt.Sprintf("error reading CSV: %s", readErr))
			return readErr
		}
		rows = append(rows, row)
	}

	// Determine output path.
	outPath := flagOutput
	if outPath == "" {
		ext := filepath.Ext(flagFile)
		base := strings.TrimSuffix(flagFile, ext)
		outPath = base + "_validated" + ext
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		output.PrintError(os.Stderr, fmt.Sprintf("could not create output file: %s", err))
		return err
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)

	// Write header with new columns.
	outHeader := append(header, "truelist_state", "truelist_sub_state", "truelist_suggestion")
	if err := writer.Write(outHeader); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Progress bar.
	bar := progressbar.NewOptions(len(rows),
		progressbar.OptionSetDescription("Validating"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100*1e6), // 100ms as nanoseconds
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetPredictTime(true),
	)

	counts := map[string]int{"valid": 0, "invalid": 0, "risky": 0, "unknown": 0}

	for _, row := range rows {
		if emailColIdx >= len(row) {
			// Row doesn't have enough columns; write it with empty validation fields.
			outRow := append(row, "", "", "")
			_ = writer.Write(outRow)
			_ = bar.Add(1)
			continue
		}

		email := strings.TrimSpace(row[emailColIdx])
		if email == "" {
			outRow := append(row, "", "", "")
			_ = writer.Write(outRow)
			_ = bar.Add(1)
			continue
		}

		result, validateErr := c.Validate(context.Background(), email)
		if validateErr != nil {
			fmt.Fprintf(os.Stderr, "\nWarning: failed to validate %s: %s\n", email, validateErr)
			outRow := append(row, "error", validateErr.Error(), "")
			_ = writer.Write(outRow)
			_ = bar.Add(1)
			continue
		}

		counts[strings.ToLower(result.State)]++

		outRow := append(row, result.State, result.SubState, result.Suggestion)
		if writeErr := writer.Write(outRow); writeErr != nil {
			return fmt.Errorf("failed to write row: %w", writeErr)
		}

		_ = bar.Add(1)
	}

	_ = bar.Finish()

	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("failed to write CSV output: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\nResults written to %s\n", outPath)

	total := counts["valid"] + counts["invalid"] + counts["risky"] + counts["unknown"]
	output.PrintSummary(os.Stderr, total, counts["valid"], counts["invalid"], counts["risky"], counts["unknown"])

	return nil
}

// findEmailColumn locates the email column in the CSV header.
// If columnName is provided, it matches exactly (case-insensitive).
// Otherwise, it auto-detects by looking for common email column names.
func findEmailColumn(header []string, columnName string) int {
	if columnName != "" {
		for i, h := range header {
			if strings.EqualFold(strings.TrimSpace(h), columnName) {
				return i
			}
		}
		return -1
	}

	candidates := []string{"email", "email_address", "emailaddress", "e-mail", "mail"}
	for i, h := range header {
		normalized := strings.ToLower(strings.TrimSpace(h))
		for _, candidate := range candidates {
			if normalized == candidate {
				return i
			}
		}
	}
	return -1
}
