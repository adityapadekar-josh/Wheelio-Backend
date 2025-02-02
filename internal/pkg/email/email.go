// internal/email/email.go
package email

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type service struct {
	APIKey    string
	FromName  string
	FromEmail string
}

type Service interface {
	SendEmail(toName, toEmail, subject, plainTextContent string) error
}

func NewService(apiKey, fromName, fromEmail string) Service {
	return &service{APIKey: apiKey, FromName: fromName, FromEmail: fromEmail}
}

func (s *service) SendEmail(toName, toEmail, subject, plainTextContent string) error {
	from := mail.NewEmail(s.FromName, s.FromEmail)
	to := mail.NewEmail(toName, toEmail)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, "")

	client := sendgrid.NewSendClient(s.APIKey)

	_, err := client.Send(message)

	if err != nil {
		return err
	}

	return nil
}
