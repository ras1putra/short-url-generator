package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetEncoder_Development(t *testing.T) {
	enc := getEncoder("development")
	assert.NotNil(t, enc)
}

func TestGetEncoder_Production(t *testing.T) {
	enc := getEncoder("production")
	assert.NotNil(t, enc)
}

func TestWithUser(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	l := WithUser("user-123")
	assert.NotNil(t, l)
}

func TestWithFields(t *testing.T) {
	_ = zap.ReplaceGlobals(zap.NewNop())
	l := WithFields(zap.String("key", "value"))
	assert.NotNil(t, l)
}

func TestInit_Development(t *testing.T) {
	err := Init("development")
	require.NoError(t, err)
	assert.NotNil(t, zap.L())
	_ = zap.ReplaceGlobals(zap.NewNop())
	os.RemoveAll("logs")
}

func TestInit_Production(t *testing.T) {
	err := Init("production")
	require.NoError(t, err)
	assert.NotNil(t, zap.L())
	_ = zap.ReplaceGlobals(zap.NewNop())
	os.RemoveAll("logs")
}