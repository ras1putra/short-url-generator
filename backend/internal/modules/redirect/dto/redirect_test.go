package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorResponse_JSON(t *testing.T) {
	resp := ErrorResponse{Error: "something went wrong", RemainingMs: 5000}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"error":"something went wrong"`)
	assert.Contains(t, string(data), `"remaining_ms":5000`)
}

func TestErrorResponse_NoRemainingMs(t *testing.T) {
	resp := ErrorResponse{Error: "not found"}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"error":"not found"`)
	assert.NotContains(t, string(data), `"remaining_ms"`)
}

func TestCompletionResponse_JSON(t *testing.T) {
	resp := CompletionResponse{Success: true, DestinationURL: "https://example.com"}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"success":true`)
	assert.Contains(t, string(data), `"destination_url":"https://example.com"`)
}

func TestCompletionResponse_Failure(t *testing.T) {
	resp := CompletionResponse{Success: false}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"success":false`)
}
