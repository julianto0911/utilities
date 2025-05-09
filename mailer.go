package utilities

import (
	"log"

	"github.com/stretchr/testify/mock"
	"gopkg.in/gomail.v2"
)

type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) SendMail(p MailContent) error {
	args := m.Called(p)
	return args.Error(0)
}

type MailerConfig struct {
	Host         string
	Port         int
	AuthEmail    string
	AuthPassword string
}

func GetMailerConfig() MailerConfig {
	cfg := MailerConfig{
		Host:         EnvString("MAILER_HOST"),
		Port:         EnvInt("MAILER_PORT"),
		AuthEmail:    EnvString("AUTH_EMAIL"),
		AuthPassword: EnvString("AUTH_PASSWORD"),
	}

	return cfg
}

func NewMailer(config MailerConfig) Mailer {
	return &mailer{
		Config: config,
	}
}

type Mailer interface {
	SendMail(c MailContent) error
}

type mailer struct {
	Config MailerConfig
}

type MailCC struct {
	Email string
	Name  string
}

type MailBody struct {
	Content     string
	ContentType string
}

func (m *mailer) doMailLog(c MailContent, err error) {
	if !c.Log {
		return
	}

	if err != nil {
		log.Println("====Email Error====")
	} else {
		log.Println("====Email Sent====")
	}

	log.Println("From :", m.Config.AuthEmail)
	log.Println("To :", c.Recipient)
	log.Println("CC :", c.CC)
	log.Println("Subject :", c.Subject)
	log.Println("Type :", c.Body.ContentType)
	log.Println("Body :", c.Body.Content)
	log.Println("Attachment :", c.Attachments)

	if err != nil {
		log.Println("Error :", err)
	}
}

type MailContent struct {
	Recipient   []string
	CC          []MailCC
	Subject     string
	Attachments []string
	Body        MailBody
	Log         bool
}

func (m *mailer) SendMail(c MailContent) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", m.Config.AuthEmail)

	for _, recipient := range c.Recipient {
		mailer.SetHeader("To", recipient)
	}

	for _, cc := range c.CC {
		mailer.SetAddressHeader("Cc", cc.Email, cc.Name)
	}

	mailer.SetHeader("Subject", c.Subject)
	mailer.SetBody(c.Body.ContentType, c.Body.Content)

	for _, attachment := range c.Attachments {
		mailer.Attach(attachment)
	}

	dialer := gomail.NewDialer(
		m.Config.Host,
		m.Config.Port,
		m.Config.AuthEmail,
		m.Config.AuthPassword,
	)

	err := dialer.DialAndSend(mailer)
	if err != nil {
		m.doMailLog(c, err)
		return err
	}

	m.doMailLog(c, nil)
	return nil
}
