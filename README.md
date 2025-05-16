# GitHub CodeQL Report Generator

A production-grade CLI tool to generate comprehensive reports from GitHub CodeQL alerts.

## Features

- Reads a CSV file containing repository and alert information
- Fetches detailed CodeQL alert data from the GitHub API
- Generates a formatted CSV report with comprehensive alert information
- Proper error handling and logging

## Installation

```bash
# Install using GH CLI
gh extension install lindluni/gh-generate-codeql-report
```

## Usage

### Command-line Options

```
gh generate-codeql-report [flags]

Flags:
  --token string     GitHub access token (required)
  --input string     Path to the input CSV file (required)
  --output string    Path to the output CSV file (default "codeql-report.csv")
  --log string       Path to the log file (default: stderr)
  --verbose          Enable verbose output
  --help             Show help information
```

### Output CSV Format

The generated report will include the following columns:
- `Org`: Organization/owner name
- `Repo`: Repository name
- `Alert ID`: The alert identifier
- `Severity`: Alert severity (high, medium, low)
- `Short Description`: Brief description of the alert
- `Full Description`: Detailed description of the alert
- `File Path`: Path to the affected file
- `Start Line`: Starting line number
- `Start Column`: Starting column number
- `End Line`: Ending line number
- `End Column`: Ending column number

## Examples

### Basic Usage

```bash
# Using command-line flags
gh generate-codeql-report --token ghp_your_token_here --input alerts.csv --output report.csv
```

### Advanced Usage

```bash
# Enable verbose output and custom log file
gh generate-codeql-report --token ghp_your_token_here --input alerts.csv --verbose --log logs/detailed.log
```

## License

MIT License
