package attacker

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync/atomic"
)

// Feeder provides a stream of data records
type Feeder interface {
	Next() map[string]string
}

// CSVFeeder reads data from a CSV file and cycles through records
type CSVFeeder struct {
	idx     uint64
	records []map[string]string
}

// NewCSVFeeder creates a new CSV feeder
func NewCSVFeeder(path string) (*CSVFeeder, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open csv file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// Read all data
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv data: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("csv file must have a header and at least one row")
	}

	headers := rows[0]
	// Validate headers
	for _, h := range headers {
		if h == "" {
			return nil, fmt.Errorf("csv header contains empty field")
		}
	}

	var records []map[string]string
	for _, row := range rows[1:] {
		record := make(map[string]string)
		for i, val := range row {
			if i < len(headers) {
				record[headers[i]] = val
			}
		}
		records = append(records, record)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("csv file contains no data rows")
	}

	return &CSVFeeder{
		idx:     0,
		records: records,
	}, nil
}

// Next returns the next record in the sequence, looping back to start if needed
func (f *CSVFeeder) Next() map[string]string {
	// Atomic increment for lock-free access
	i := atomic.AddUint64(&f.idx, 1) - 1
	return f.records[i%uint64(len(f.records))]
}
