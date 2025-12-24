// Package main contains the consumer logic for sending emails to recipients.
package main

import (
	"fmt"
	"net/smtp"
	"sync"
	"time"
)

// emailWorker processes recipients from a channel and sends personalized emails.
// Each worker runs concurrently and continues until the channel is closed.
// The worker executes the email template for each recipient and sends via SMTP.
// Failed emails are sent to the DLQ channel for tracking and potential retry.
func emailWorker(id int, ch chan Recipient, dlqCh chan FailedEmail, wg *sync.WaitGroup) {
	defer wg.Done()

	// Process recipients until the channel is closed
	for recipient := range ch {
		// Configure SMTP server connection (using local mailhog for development)
		smtpHost := "localhost"
		smtpPort := "1025"

		// Generate personalized email content from template
		msg, err := executeTemplate(recipient)
		if err != nil {
			fmt.Printf("Worker %d: Error parsing template for %s\n", id, recipient.Email)
			// Add failed email to dead letter queue (DLQ)
			dlqCh <- FailedEmail{
				Recipient: recipient,
				Error:     fmt.Sprintf("Template parsing error: %v", err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
			continue
		}

		fmt.Printf("Worker %d: Sending email to %s \n", id, recipient.Email)

		// Send email via SMTP
		err = smtp.SendMail(smtpHost+":"+smtpPort, nil, "khanshariq92213@gmail.com", []string{recipient.Email}, []byte(msg))
		if err != nil {
			fmt.Printf("Worker %d: Failed to send email to %s: %v\n", id, recipient.Email, err)
			// Add failed email to dead letter queue (DLQ)
			dlqCh <- FailedEmail{
				Recipient: recipient,
				Error:     fmt.Sprintf("SMTP error: %v", err),
				Timestamp: time.Now().Format(time.RFC3339),
			}
			continue
		}

		// Throttle email sending to avoid overwhelming the SMTP server
		time.Sleep(50 * time.Millisecond)

		fmt.Printf("Worker %d: Sent email to %s \n", id, recipient.Email)

	}
}
