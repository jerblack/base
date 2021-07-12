package mail

import (
	"github.com/go-gomail/gomail"
	"os"
	"strconv"
	"strings"
)

type Email struct {
	To          string
	Subject     string
	Body        string
	Attachments []string
}

func (e *Email) Send() error {
	if e.To == "" {
		e.To = os.Getenv("ALERT_EMAIL")
	}
	b, err := os.ReadFile("/etc/sendmail")
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	server, portStr, user, pass := lines[0], lines[1], lines[2], lines[3]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", e.To)
	m.SetHeader("To", e.To)
	m.SetHeader("Subject", e.Subject)
	m.SetBody("text/plain", e.Body)
	for _, attach := range e.Attachments {
		m.Attach(attach)
	}
	d := gomail.NewDialer(server, port, user, pass)
	err = d.DialAndSend(m)
	return err
}
