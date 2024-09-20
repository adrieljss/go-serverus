package email

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/adrieljansen/go-serverus/db"
	"github.com/adrieljansen/go-serverus/env"
	"github.com/adrieljansen/go-serverus/result"
	"github.com/adrieljansen/go-serverus/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// email module for GoServerus
// tested for GMail

type PendingConfirmationEmail struct {
	VerificationCode string
	UserToCreate     db.RequiredUser
}

var PendingConfirmationEmailRegisterCache *utils.TtlMap[string, PendingConfirmationEmail]
var emailVerificationHtmlCache string

// loads all neccessary files to the cache
func StartEmailService() {
	b, err := os.ReadFile("email/verification.html")
	if err != nil {
		logrus.Fatal("cannot load/read email verification file to cache")
		return
	}
	logrus.Warn("successfully loaded email templates to cache")
	emailVerificationHtmlCache = string(b)

	// every 10 minutes clear expired
	PendingConfirmationEmailRegisterCache = utils.NewTtlMap[string, PendingConfirmationEmail](time.Minute * 10)
	logrus.Warn("started goroutine for confirmation email ttl cache")
}

type EmailVerificationArgs struct {
	AppName         string
	AppRootUrl      string
	VerifyEmailCode string
}

// Sends a verification link of /verifyEmail/{emailVerificationUrlCode} to the target email
func SendEmailVerification(targetEmail string, emailVerificationUrlCode string) *result.Error {
	m := gomail.NewMessage()
	m.SetHeader("From", env.CSMTPFrom)
	m.SetHeader("To", targetEmail)
	m.SetHeader("Subject", fmt.Sprintf("[%s] Email Verification", env.CAppName))
	template := template.New("email template")
	template, err := template.Parse(emailVerificationHtmlCache)
	if err != nil {
		logrus.Fatal("cannot format email verification html file")
		return nil
	}
	var buf bytes.Buffer
	args := EmailVerificationArgs{
		AppName:         env.CAppName,
		AppRootUrl:      env.CAppRootUrl,
		VerifyEmailCode: emailVerificationUrlCode,
	}
	e := template.Execute(&buf, args)
	if e != nil {
		logrus.Fatal("cannot [execute] format email verification html file")
		return nil
	}
	m.SetBody("text/html", buf.String())
	d := gomail.NewDialer(env.CSMTPHost, int(env.CSMTPPort), env.CSMTPFrom, env.CSMTPPass)
	if ee := d.DialAndSend(m); ee != nil {
		return result.Err(404, ee, "CANT_SEND_EMAIL", "unable to send email to target email")
	}
	return nil
}

func ResetPasswordEmail() {

}
