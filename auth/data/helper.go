package data

import (
	"log"
	"net/smtp"
	"os"

	"github.com/google/uuid"
)

func SendEmail(providedEmail, activationUUID, intention string) (string, error) {
	const accountActivationPath = "https://localhost/api/auths/activate/"
	const accountRecoveryPath = "https://localhost:4200/recover-account"
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

	if intention == "activation" {
		subject = "StayInn Account Activation"
		body = "Follow the verification link to activate your StayInn account: \n" + accountActivationPath + activationUUID
	} else if intention == "recovery" {
		subject = "StayInn Password Recovery"
		body = "To reset your password, copy the given code & then follow the recovery link: \n" + activationUUID + "\n" + accountRecoveryPath
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
		return "", err
	}
	log.Println("Mail successfully sent")
	return activationUUID, nil
}

func generateActivationUUID() string {
	requestUUID := uuid.New().String()
	return requestUUID
}
