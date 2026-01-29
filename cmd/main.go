package main

import (
	"log"

	"fake-mail/internal/smtp"
	"fake-mail/internal/store"
	"fake-mail/internal/web"
)

func main() {
	// Create shared email store
	emailStore := store.New()

	// Start SMTP server
	smtpServer := smtp.New(":1025", emailStore)
	go func() {
		if err := smtpServer.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	// Start web server
	webServer := web.New(":8025", emailStore)
	if err := webServer.Start(); err != nil {
		log.Fatal(err)
	}
}
