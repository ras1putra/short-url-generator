package mailer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	m := New("test-key", "from@test.com", "http://localhost:3000")
	assert.NotNil(t, m)
	assert.Equal(t, "from@test.com", m.from)
	assert.Equal(t, "http://localhost:3000", m.frontendURL)
}

func TestRenderVerificationTemplate(t *testing.T) {
	m := New("", "noreply@test.com", "http://localhost:3000")
	html, err := m.renderTemplate(verificationTemplate, struct {
		Name string
		Link string
	}{Name: "John", Link: "http://localhost:3000/verify-email?token=abc123"})
	assert.NoError(t, err)
	assert.Contains(t, html, "John")
	assert.Contains(t, html, "VERIFY YOUR EMAIL")
	assert.Contains(t, html, "abc123")
}

func TestRenderPasswordResetTemplate(t *testing.T) {
	m := New("", "noreply@test.com", "http://localhost:3000")
	html, err := m.renderTemplate(passwordResetTemplate, struct {
		Name string
		Link string
	}{Name: "Jane", Link: "http://localhost:3000/reset-password?token=xyz789"})
	assert.NoError(t, err)
	assert.Contains(t, html, "Jane")
	assert.Contains(t, html, "RESET YOUR PASSWORD")
	assert.Contains(t, html, "xyz789")
}

func TestRenderTemplate_WithSpecialChars(t *testing.T) {
	m := New("", "noreply@test.com", "http://localhost:3000")
	html, err := m.renderTemplate(verificationTemplate, struct {
		Name string
		Link string
	}{Name: "John & Doe <test>", Link: "http://localhost:3000/verify?token=a&b=c"})
	assert.NoError(t, err)
	assert.Contains(t, html, "John &amp; Doe &lt;test&gt;")
}

func TestRenderTemplate_EmptyName(t *testing.T) {
	m := New("", "noreply@test.com", "http://localhost:3000")
	html, err := m.renderTemplate(verificationTemplate, struct {
		Name string
		Link string
	}{Name: "", Link: "http://localhost:3000/verify?token=empty"})
	assert.NoError(t, err)
	assert.NotContains(t, html, "{{")
	assert.True(t, strings.HasPrefix(html, "<"))
}

func TestSendVerificationEmail_FailsWithoutClient(t *testing.T) {
	m := New("invalid-key", "from@test.com", "http://localhost:3000")
	err := m.SendVerificationEmail("test@example.com", "Test", "token123")
	assert.Error(t, err)
}

func TestSendPasswordResetEmail_FailsWithoutClient(t *testing.T) {
	m := New("invalid-key", "from@test.com", "http://localhost:3000")
	err := m.SendPasswordResetEmail("test@example.com", "Test", "token123")
	assert.Error(t, err)
}
