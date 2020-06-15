package main

import (
	"gopkg.in/gomail.v2"
)

// Subscribe info
type Subscribe struct {
	Sender         string
	Password       string
	SMTPServer     string
	SMTPServerPort int
	Subscriber     string
}

// Attachment contain attachment file path and filename when sending mail
type Attachment struct {
	FilePath string
	Filename string
}

// Mail content to subscriber
func Mail(s *Subscribe, subject string, content string, a *Attachment) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.Sender)
	m.SetHeader("To", s.Subscriber)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", content)
	if a != nil {
		m.Attach(a.FilePath, gomail.Rename(a.Filename))
	}

	d := gomail.NewDialer(s.SMTPServer, s.SMTPServerPort, s.Sender, s.Password)

	return d.DialAndSend(m)
}
