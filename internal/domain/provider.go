package domain

import "context"

// SMSProvider defines the interface for SMS/MMS providers
type SMSProvider interface {
	SendSMS(ctx context.Context, from, to, body string) error
	SendMMS(ctx context.Context, from, to, body string, attachments []string) error
}

// EmailProvider defines the interface for email providers
type EmailProvider interface {
	SendEmail(ctx context.Context, from, to, body string, attachments []string) error
}
