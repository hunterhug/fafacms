package mail

import (
	"crypto/tls"
	"gopkg.in/gomail.v2"
)

var Debug = false

type Sender struct {
	//Bcc      string
	//BccName  string
	Host     string `yaml:"Host"`
	Port     int    `yaml:"Port"`
	Account  string `yaml:"Account"`
	Password string `yaml:"Password"`
	Subject  string `yaml:"Subject"`
	From     string `yaml:"From"`
	Body     string `yaml:"Body"`
}

type Message struct {
	To     string
	ToName string
	Sender
}

func (mm *Message) Sent() error {
	if Debug {
		return nil
	}
	m := gomail.NewMessage()
	m.SetAddressHeader("From", mm.Account, mm.From)
	m.SetAddressHeader("To", mm.To, mm.ToName)
	m.SetHeader("Subject", mm.Subject)

	//m.SetHeader("Bcc",
	//	m.FormatAddress(mm.Bcc, mm.BccName))

	m.SetBody("text/html", mm.Body)

	d := gomail.NewDialer(mm.Host, mm.Port, mm.Account, mm.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
