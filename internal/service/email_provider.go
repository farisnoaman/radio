package service

import (
	"fmt"
	"net/smtp"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SMTPEmailProvider struct {
	config SMTPConfig
}

func NewSMTPEmailProvider(config SMTPConfig) *SMTPEmailProvider {
	return &SMTPEmailProvider{config: config}
}

func (p *SMTPEmailProvider) SendEmail(to, subject, body string) error {
	if p.config.Host == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		p.config.From, to, subject, body)

	var auth smtp.Auth
	if p.config.Username != "" {
		auth = smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)
	}

	return smtp.SendMail(addr, auth, p.config.From, []string{to}, []byte(msg))
}
