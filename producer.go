// Package main contains the producer logic for loading email recipients.
package main

import (
	"encoding/csv"
	"os"
)

// loadRecipient reads email recipients from a CSV file and sends them to the provided channel.
// The CSV file is expected to have a header row which is skipped.
// The channel is closed after all recipients are loaded.
// Returns an error if the file cannot be opened or read.
func loadRecipient(filePath string, ch chan Recipient) error {
	defer close(ch)

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create CSV reader and read all records
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	// Skip header row (records[1:]) and send each recipient to the channel
	for _, record := range records[1:] {
		ch <- Recipient{
			Name:  record[0],
			Email: record[1],
		}
	}

	return nil
}
