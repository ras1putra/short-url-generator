package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleUserInfo_JSON(t *testing.T) {
	info := GoogleUserInfo{
		ID:            "12345",
		Email:         "user@gmail.com",
		VerifiedEmail: true,
		Name:          "John Doe",
		Picture:       "https://example.com/pic.jpg",
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded GoogleUserInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, info.ID, decoded.ID)
	assert.Equal(t, info.Email, decoded.Email)
	assert.Equal(t, info.Name, decoded.Name)
	assert.Equal(t, info.Picture, decoded.Picture)
	assert.True(t, decoded.VerifiedEmail)
}

func TestGoogleUserInfo_EmptyFields(t *testing.T) {
	info := GoogleUserInfo{}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded GoogleUserInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Empty(t, decoded.ID)
	assert.Empty(t, decoded.Email)
	assert.Empty(t, decoded.Name)
	assert.False(t, decoded.VerifiedEmail)
}

func TestGoogleUserInfo_UnverifiedEmail(t *testing.T) {
	info := GoogleUserInfo{
		ID:            "abc",
		Email:         "test@example.com",
		VerifiedEmail: false,
		Name:          "Test",
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded GoogleUserInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.False(t, decoded.VerifiedEmail)
}
