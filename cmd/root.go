package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/lindluni/gh-generate-codeql-report/pkg/codeql"
	csvpkg "github.com/lindluni/gh-generate-codeql-report/pkg/csv"
)

var (
	// Global flags
	token      string
	inputFile  string
	outputFile string
	logFile    string
	verbose    bool

	// Logger for the application
	logger *log.Logger
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gh-generate-codeql-report",
	Short: "Generate a CodeQL report from GitHub alerts",
	Long: `Generate a comprehensive CodeQL report from GitHub alerts.
This tool takes a CSV file with repository and alert information, queries
the GitHub API for detailed information, and outputs a formatted report.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogging()
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := validateFlags(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Process alerts and generate report
		if err := generateReport(ctx); err != nil {
			logger.Printf("Error generating report: %v", err)
			fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
			os.Exit(1)
		}

		logger.Printf("Report successfully generated at %s", outputFile)
		if verbose {
			fmt.Printf("Report successfully generated at %s\n", outputFile)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	// Define flags and their default values
	RootCmd.PersistentFlags().StringVar(&token, "token", "", "GitHub access token (required)")
	RootCmd.PersistentFlags().StringVar(&inputFile, "input", "", "Path to the input CSV file (required)")
	RootCmd.PersistentFlags().StringVar(&outputFile, "output", "codeql-report.csv", "Path to the output CSV file")
	RootCmd.PersistentFlags().StringVar(&logFile, "log", "", "Path to the log file (default: stderr)")
	RootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
}

// setupLogging configures the application logger
func setupLogging() {
	var logWriter *os.File
	var err error

	if logFile == "" {
		logWriter = os.Stderr
	} else {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
			os.Exit(1)
		}

		logWriter, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
			os.Exit(1)
		}
	}

	logger = log.New(logWriter, "", log.LstdFlags|log.Lshortfile)
	logger.Println("Starting gh-generate-codeql-report")
}

// validateFlags checks if required flags are provided
func validateFlags() error {
	missing := false
	var missingFlags []string

	if token == "" {
		missing = true
		missingFlags = append(missingFlags, "token")
	}

	if inputFile == "" {
		missing = true
		missingFlags = append(missingFlags, "input")
	}

	if missing {
		return fmt.Errorf("required flag(s) not provided: %s", strings.Join(missingFlags, ", "))
	}

	return nil
}

// generateReport processes the input CSV and generates the CodeQL report
func generateReport(ctx context.Context) error {
	logger.Printf("Reading input from %s", inputFile)

	// Read input CSV
	csvReader := csvpkg.NewReader(inputFile)
	records, err := csvReader.ReadAllWithHeaders()
	if err != nil {
		return fmt.Errorf("failed to read input CSV: %w", err)
	}

	logger.Printf("Found %d records to process", len(records))

	// Initialize CodeQL client
	client := codeql.NewClient(token, logger)

	// Process each alert
	var alerts []codeql.Alert
	var processErrors int

	// Set up output headers
	outputHeaders := []string{
		"Org", "Repo", "Alert ID", "Severity",
		"Short Description", "Full Description",
		"File Path", "Start Line", "Start Column",
		"End Line", "End Column",
	}

	// Create data for writing
	var csvData [][]string

	for i, record := range records {
		if verbose {
			fmt.Printf("Processing record %d/%d\n", i+1, len(records))
		}

		// Extract repository owner and name
		repoFullName := record["Repository"]
		repoParts := strings.Split(repoFullName, "/")
		if len(repoParts) != 2 {
			logger.Printf("Invalid repository format: %s", repoFullName)
			processErrors++
			continue
		}

		owner := repoParts[0]
		repo := repoParts[1]

		// Parse alert number
		alertNumber := record["Alert Number"]
		alertNumberInt, err := strconv.ParseInt(alertNumber, 10, 64)
		if err != nil {
			logger.Printf("Failed to parse alert number '%s': %v", alertNumber, err)
			processErrors++
			continue
		}

		// Get alert details
		alert, err := client.GetAlert(ctx, owner, repo, alertNumberInt)
		if err != nil {
			logger.Printf("Failed to get alert #%s for %s: %v", alertNumber, repoFullName, err)
			processErrors++
			continue
		}

		alerts = append(alerts, *alert)

		// Add to CSV data
		csvData = append(csvData, []string{
			alert.Owner,
			alert.Repo,
			strconv.Itoa(alert.ID),
			alert.Severity,
			alert.ShortDesc,
			alert.FullDesc,
			alert.FilePath,
			strconv.Itoa(alert.StartLine),
			strconv.Itoa(alert.StartColumn),
			strconv.Itoa(alert.EndLine),
			strconv.Itoa(alert.EndColumn),
		})
	}

	logger.Printf("Successfully processed %d/%d alerts", len(alerts), len(records))
	if processErrors > 0 {
		logger.Printf("Failed to process %d alerts", processErrors)
	}

	// Write output CSV
	writer := csvpkg.NewWriter(outputFile, outputHeaders)
	if err := writer.WriteAll(csvData); err != nil {
		return fmt.Errorf("failed to write output CSV: %w", err)
	}

	return nil
}
