package mailer

import (
	"context"
	"email_svc/internal/model"
	"errors"
	"fmt"
	mailersend "github.com/mailersend/mailersend-go"
	"log"
	"strings"
	"time"
)

type Mailer struct {
	client *mailersend.Mailersend
}

func NewMailer(client *mailersend.Mailersend) *Mailer {
	return &Mailer{client: client}
}

const (
	SenderName        = "DuelJack"
	SenderEmail       = "MS_UJZsWa@test-zkq340epjzkgd796.mlsender.net"
	defaultRetryDelay = 2 * time.Second
	maxRetries        = 1
	emailSendTimeout  = 10 * time.Second
)

func (m *Mailer) Send(ctx context.Context, detail model.EmailSentDetail) error {
	operationCtx, operationCancel := context.WithTimeout(ctx, emailSendTimeout)
	defer operationCancel()
	if detail.To == "" {
		log.Printf("Error: Recipient email address is empty for subject: %s. Skipping send.", detail.Subject)
		return errors.New("email subject is empty")
	}
	from := mailersend.From{
		Name:  SenderName,
		Email: SenderEmail,
	}

	recipients := []mailersend.Recipient{
		{Email: detail.To},
	}

	message := m.client.Email.NewMessage()
	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(detail.Subject)
	message.SetText(detail.Body)
	message.SetHTML("")

	tagFromSubject := strings.ToLower(strings.ReplaceAll(detail.Subject, " ", "-"))
	tags := []string{"transactional", tagFromSubject}
	if len(tagFromSubject) > 0 && len(tagFromSubject) < 64 {
		message.SetTags(tags)
	} else {
		message.SetTags([]string{"transactional"})
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retrying email to %s (attempt %d/%d)...", detail.To, attempt+1, maxRetries+1)
			select {
			case <-time.After(defaultRetryDelay):
			case <-operationCtx.Done():
				log.Printf("Email send retry for %s cancelled due to context deadline: %v", detail.To, operationCtx.Err())
				if lastErr != nil {
					return fmt.Errorf("retry for %s cancelled by context deadline, last error: %w", detail.To, lastErr)
				}
				return fmt.Errorf("retry for %s cancelled by context deadline: %w", detail.To, operationCtx.Err())
			}
		}

		log.Printf("Attempting to send email to %s (attempt %d of %d)", detail.To, attempt+1, maxRetries+1)
		res, err := m.client.Email.Send(operationCtx, message)

		var messageID string
		if res != nil && res.Header != nil {
			if msgIDHeaders := res.Header.Values("X-Message-Id"); len(msgIDHeaders) > 0 {
				messageID = msgIDHeaders[0]
			}
		}

		apiStatus := ""
		if res != nil {
			apiStatus = res.Status
		}

		if err == nil {
			log.Printf("Email to %s sent successfully. Message ID: %s, API Status: %s", detail.To, messageID, apiStatus)
			return nil
		}

		// Error occurred
		lastErr = err
		log.Printf("MailerSend Send (attempt %d) to %s failed: %v | API Status: %s | Message ID (if any): %s", attempt+1, detail.To, err, apiStatus, messageID)
	}

	log.Printf("All %d attempts to send email to %s failed.", maxRetries+1, detail.To)
	return fmt.Errorf("failed to send email to %s after %d attempts: %w", detail.To, maxRetries+1, lastErr)
}
