package email

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	"github.com/adrieljansen/go-serverus/env"
	"github.com/adrieljansen/go-serverus/result"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// email module for GoServerus
// tested for GMail

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
}

type EmailVerificationArgs struct {
	AppName         string
	AppRootUrl      string
	VerifyEmailCode string
}

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
