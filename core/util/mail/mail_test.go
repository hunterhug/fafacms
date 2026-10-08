package mail

import (
	"fmt"
	"testing"
)

func TestMail_Sent(t *testing.T) {
	s := Sender{}
	s.Host = "smtp-mail.outlook.com"
	s.Port = 587
	s.Account = "your_email@example.com"
	s.Password = "ddd"

	m := new(Message)
	m.Sender = s
	m.To = "test@example.com"
	m.ToName = "user"
	m.Subject = "register"
	m.Body = "ddddddddd"

	err := m.Sent()
	if err != nil {
		fmt.Println(err.Error())
	}
}
