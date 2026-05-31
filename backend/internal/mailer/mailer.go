package mailer

import (
	"bytes"
	"html/template"

	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

type Mailer struct {
	client      *resend.Client
	from        string
	frontendURL string
}

func New(apiKey, from, frontendURL string) *Mailer {
	return &Mailer{
		client:      resend.NewClient(apiKey),
		from:        from,
		frontendURL: frontendURL,
	}
}

func (m *Mailer) SendVerificationEmail(to, name, token string) error {
	link := m.frontendURL + "/verify-email?token=" + token

	html, err := m.renderTemplate(verificationTemplate, struct {
		Name string
		Link string
	}{Name: name, Link: link})
	if err != nil {
		zap.L().Error("Failed to render verification email template", zap.Error(err))
		return err
	}

	params := &resend.SendEmailRequest{
		From:    m.from,
		To:      []string{to},
		Subject: "Verify your email",
		Html:    html,
	}

	if _, err := m.client.Emails.Send(params); err != nil {
		zap.L().Error("Failed to send verification email via Resend",
			zap.String("to", to),
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("Verification email sent via Resend",
		zap.String("to", to),
	)

	return nil
}

func (m *Mailer) SendPasswordResetEmail(to, name, token string) error {
	link := m.frontendURL + "/reset-password?token=" + token

	html, err := m.renderTemplate(passwordResetTemplate, struct {
		Name string
		Link string
	}{Name: name, Link: link})
	if err != nil {
		zap.L().Error("Failed to render password reset email template", zap.Error(err))
		return err
	}

	params := &resend.SendEmailRequest{
		From:    m.from,
		To:      []string{to},
		Subject: "Reset your password",
		Html:    html,
	}

	if _, err := m.client.Emails.Send(params); err != nil {
		zap.L().Error("Failed to send password reset email via Resend",
			zap.String("to", to),
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("Password reset email sent via Resend",
		zap.String("to", to),
	)

	return nil
}

func (m *Mailer) renderTemplate(tmplStr string, data interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
