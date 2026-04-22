package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// Row is a map from normalized column name to cell value.
type Row map[string]string

// Parse reads a CSV from r and returns one Row per data line.
// Headers are normalized to lowercase, trimmed, and spaces replaced with underscores.
func Parse(r io.Reader) ([]Row, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}
	for i, h := range headers {
		headers[i] = normalizeHeader(h)
	}

	var rows []Row
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading CSV row: %w", err)
		}
		row := make(Row, len(headers))
		for i, h := range headers {
			if i < len(record) {
				row[h] = strings.TrimSpace(record[i])
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func normalizeHeader(h string) string {
	h = strings.TrimSpace(h)
	h = strings.ToLower(h)
	h = strings.ReplaceAll(h, " ", "_")
	h = strings.ReplaceAll(h, "-", "_")
	return h
}
