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
	"github.com/pquerna/otp/totp"
	"github.com/sirupsen/logrus"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"gopkg.in/gomail.v2"
)

// email module for GoServerus
// tested for GMail

type PendingConfirmationEmail struct {
	VerificationCode string
	UserToCreate     *db.RequiredUser
}

type PendingResetPass struct {
	OTPCode     string
	User        *db.User
	NewPassword string
}

var PendingConfirmationEmailRegisterCache *utils.TtlMap[string, PendingConfirmationEmail]
var PendingResetPassCache *utils.TtlMap[string, PendingResetPass]
var emailVerificationTemplate *template.Template
var resetPassTemplate *template.Template
var emailDialer *gomail.Dialer

func bindToTemplate(filePath string, bindTo **template.Template) {
	b, err := os.ReadFile(filePath)

	if err != nil {
		logrus.Fatalf("cannot load file %s in email module", filePath)
		return
	}

	mnf := minify.New()
	mnf.AddFunc("text/html", html.Minify)

	minified, err := mnf.Bytes("text/html", b)
	if err != nil {
		logrus.Fatalf("cannot compress/minify file %s in email module", filePath)
		return
	}

	// template name as file path
	tmp := template.New(filePath)
	tmp, err = tmp.Parse(string(minified))
	if err != nil {
		logrus.Fatalf("cannot pass file %s as email template", filePath)
		return
	}

	*bindTo = tmp
	logrus.Warnf("loaded email template %s", filePath)
}

// loads all neccessary files to the cache
func StartEmailService() {
	bindToTemplate("email/verification.html", &emailVerificationTemplate)
	bindToTemplate("email/resetpass.html", &resetPassTemplate)

	// every 2 hours clear expired to free up cache
	PendingConfirmationEmailRegisterCache = utils.NewTtlMap[string, PendingConfirmationEmail](env.EmailConfirmationObliteratorInterval)
	PendingResetPassCache = utils.NewTtlMap[string, PendingResetPass](env.EmailResetPassObliteratorInterval)
	logrus.Warn("started goroutine for confirmation email and reset pass ttl cache")

	emailDialer = gomail.NewDialer(env.CSMTPHost, int(env.CSMTPPort), env.CSMTPFrom, env.CSMTPPass)
}

// generate an otp code with `secret` and launch a goroutine of cbfn
func GenerateOtpAndAct(secret string, cbfn func(string)) *result.Error {
	otpCode, errOtp := totp.GenerateCode(secret, time.Now())
	if errOtp != nil {
		return result.ServerErr(errOtp)
	}

	go func() {
		cbfn(otpCode)
	}()
	return nil
}

type EmailVerificationArgs struct {
	AppName         string
	AppRootUrl      string
	UserId          string
	VerifyEmailCode string
}

type ResetPassArgs struct {
	AppName       string
	AppRootUrl    string
	UserId        string
	ResetPassCode string
}

// Sends a otp code for email verification to the target email
func SendEmailVerification(targetEmail string, emailVerificationUrlCode string, userId string) *result.Error {
	m := gomail.NewMessage()
	m.SetHeader("From", env.CSMTPFrom)
	m.SetHeader("To", targetEmail)
	m.SetHeader("Subject", fmt.Sprintf("[%s] Email Verification", env.CAppName))
	var buf bytes.Buffer
	args := EmailVerificationArgs{
		AppName:         env.CAppName,
		AppRootUrl:      env.CAppRootUrl,
		VerifyEmailCode: emailVerificationUrlCode,
		UserId:          userId,
	}
	e := emailVerificationTemplate.Execute(&buf, args)
	if e != nil {
		logrus.Fatal("cannot [execute] format email verification html file")
		return nil
	}
	m.SetBody("text/html", buf.String())
	m.Embed("./public/logo.png") // cop can not be svg
	if ee := emailDialer.DialAndSend(m); ee != nil {
		return result.Err(404, ee, "CANT_SEND_EMAIL", "unable to send email to target email")
	}
	return nil
}

// sends an otp code for reset password to the target email
func SendEmailResetPassword(targetEmail string, resetPassCode string, userId string) *result.Error {
	m := gomail.NewMessage()
	m.SetHeader("From", env.CSMTPFrom)
	m.SetHeader("To", targetEmail)
	m.SetHeader("Subject", fmt.Sprintf("[%s] Reset Pass OTP", env.CAppName))
	var buf bytes.Buffer
	args := ResetPassArgs{
		AppName:       env.CAppName,
		AppRootUrl:    env.CAppRootUrl,
		ResetPassCode: resetPassCode,
		UserId:        userId,
	}
	e := resetPassTemplate.Execute(&buf, args)
	if e != nil {
		logrus.Fatal("cannot [execute] format reset password html file")
		return nil
	}
	m.SetBody("text/html", buf.String())
	m.Embed("./public/logo.png") // cop can not be svg
	if ee := emailDialer.DialAndSend(m); ee != nil {
		return result.Err(404, ee, "CANT_SEND_EMAIL", "unable to send email to target email")
	}
	return nil
}
