package email

import (
	"log/slog"

	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/config"
	"github.com/adityapadekar-josh/Wheelio-Backend.git/internal/pkg/apperrors"
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

func NewService() Service {
	cfg := config.GetConfig()
	return &service{APIKey: cfg.EmailService.ApiKey, FromName: cfg.EmailService.FromName, FromEmail: cfg.EmailService.FromEmail}
}

func (s *service) SendEmail(toName, toEmail, subject, plainTextContent string) error {
	from := mail.NewEmail(s.FromName, s.FromEmail)
	to := mail.NewEmail(toName, toEmail)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, "")

	client := sendgrid.NewSendClient(s.APIKey)

	_, err := client.Send(message)
	if err != nil {
		slog.Error("failed to send email", "error", err)
		return apperrors.ErrEmailSendFailed
	}

	return nil
}
