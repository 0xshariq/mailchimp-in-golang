// Package main implements a concurrent email campaign sender using worker pools.
// It reads recipients from a CSV file and sends personalized emails using templates.
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
)

// Recipient represents an email recipient with their personal information.
type Recipient struct {
	Name  string // Name of the recipient
	Email string // Email address of the recipient
}

// FailedEmail represents an email that failed to send, along with error information.
type FailedEmail struct {
	Recipient Recipient
	Error     string
	Timestamp string
}

// Dead Letter Queue (DLQ) for failed emails
var (
	dlq      []FailedEmail
	dlqMutex sync.Mutex
)

// main initializes the email campaign by setting up a channel for recipients,
// loading recipients from CSV, and spawning worker goroutines to send emails concurrently.
func main() {
	recipientChannel := make(chan Recipient)
	dlqChannel := make(chan FailedEmail, 100)

	go func() {
		loadRecipient("./emails.csv", recipientChannel)
	}()

	// DLQ collector goroutine
	go func() {
		for failed := range dlqChannel {
			dlqMutex.Lock()
			dlq = append(dlq, failed)
			dlqMutex.Unlock()
		}
	}()

	var wg sync.WaitGroup
	workerCount := 5

	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go emailWorker(i, recipientChannel, dlqChannel, &wg)
	}

	wg.Wait()
	close(dlqChannel)

	// Print DLQ summary
	printDLQSummary()
}

// executeTemplate parses and executes an email template with the given recipient's data.
// It returns the generated email content as a string or an error if template processing fails.
func executeTemplate(r Recipient) (string, error) {
	t, err := template.ParseFiles("./email.tmpl")
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer

	err = t.Execute(&tpl, r)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}

// printDLQSummary displays all failed emails from the dead letter queue.
func printDLQSummary() {
	dlqMutex.Lock()
	defer dlqMutex.Unlock()

	if len(dlq) == 0 {
		fmt.Println("\n✓ All emails sent successfully!")
		return
	}

	fmt.Printf("\n⚠ Failed to send %d email(s):\n", len(dlq))
	fmt.Println("----------------------------------------")
	for i, failed := range dlq {
		fmt.Printf("%d. Email: %s, Name: %s\n", i+1, failed.Recipient.Email, failed.Recipient.Name)
		fmt.Printf("   Error: %s\n", failed.Error)
		fmt.Printf("   Time: %s\n", failed.Timestamp)
		fmt.Println()
	}
}
