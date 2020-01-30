package utils

import (
	"log"
	"net/smtp"
	"os"
)

func SendEmail(content,subject string) {
	from := os.Getenv("MAIL_USER")
	pass := os.Getenv("MAIL_PASS")
	to := os.Getenv("MAIL_TO")
	auth := smtp.PlainAuth("smtp.gmail.com:587", from, pass, "smtp.gmail.com")

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: "+subject+"\n\n" +
		content

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}
