package notify

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type EmailChannel struct {
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
}

func NewEmailChannel(smtpHost, smtpPort, username, password, from string) *EmailChannel {
	return &EmailChannel{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		from:     from,
	}
}

func (c *EmailChannel) Send(ctx context.Context, to, subject, content string) error {
_addr := fmt.Sprintf("%s:%d", c.smtpHost, c.smtpPort)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		c.from, to, subject, content)

	auth := smtp.PlainAuth("", c.username, c.password, c.smtpHost)

	err := smtp.SendMail(_addr, auth, c.from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("smtp send failed: %w", err)
	}
	return nil
}
