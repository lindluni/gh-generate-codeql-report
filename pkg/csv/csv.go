// Package csv provides functionality for CSV file operations.
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// Reader handles reading and parsing CSV files.
type Reader struct {
	filePath string
}

// NewReader creates a new CSV reader for the specified file.
func NewReader(filePath string) *Reader {
	return &Reader{
		filePath: filePath,
	}
}

// ReadAllWithHeaders reads all records from a CSV file and returns them as a slice of maps.
// Each map represents a row, with keys being the column headers.
func (r *Reader) ReadAllWithHeaders() ([]map[string]string, error) {
	f, err := os.Open(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", r.filePath, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// Read headers
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	var records []map[string]string

	// Read rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		if len(row) != len(headers) {
			return nil, fmt.Errorf("row length (%d) does not match header length (%d): %v", len(row), len(headers), row)
		}

		// Build map for this row
		rowMap := make(map[string]string)
		for i, header := range headers {
			rowMap[header] = row[i]
		}
		records = append(records, rowMap)
	}

	return records, nil
}

// Writer handles writing CSV data to files.
type Writer struct {
	filePath string
	headers  []string
}

// NewWriter creates a new CSV writer for the specified file.
func NewWriter(filePath string, headers []string) *Writer {
	return &Writer{
		filePath: filePath,
		headers:  headers,
	}
}

// WriteAll writes all records to a CSV file.
func (w *Writer) WriteAll(records [][]string) error {
	f, err := os.Create(w.filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", w.filePath, err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(w.headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write records
	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("failed to write CSV records: %w", err)
	}

	return nil
}
