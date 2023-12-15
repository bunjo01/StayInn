package data

import (
	"log"
	"net/smtp"
	"os"
)

func SendNotificationEmail(providedEmail, intention string) (bool, error) {
	// Sender data
	from := os.Getenv("MAIL_ADDRESS")
	password := os.Getenv("MAIL_PASSWORD")

	// Receiver email
	to := []string{
		providedEmail,
	}

	// smtp server config
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	address := smtpHost + ":" + smtpPort
	var subject string
	var body string

	if intention == "reservation-new" {
		subject = "StayInn Notification - New Reservation"
		body = "New reservation was made for your accommodation. Login to StayInn to check it out!"
	} else if intention == "reservation-deleted" {
		subject = "StayInn Notification - Reservation canceled"
		body = "An user has deleted the reservation for your accommodation. Login to StayInn to see the details."
	} else if intention == "rating-host" {
		subject = "StayInn Notification - New Host Rating"
		body = "An user has rated you. Login to StayInn to see the details."
	} else if intention == "rating-accommodation" {
		subject = "StayInn Notification - New Accommodation Rating"
		body = "An user has rated your accommodation. Login to StayInn to see the details."
	}

	// Text
	stringMsg :=
		"From: " + from + "\n" +
			"To: " + to[0] + "\n" +
			"Subject: " + subject + "\n\n" +
			body

	message := []byte(stringMsg)

	// Email Sender Auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(address, auth, from, to, message)
	if err != nil {
		log.Println("Error sending mail", err)
		return false, err
	}

	log.Println("Mail successfully sent")

	return true, nil
}
