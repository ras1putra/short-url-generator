package dto

type ErrorResponse struct {
	Error       string `json:"error"`
	RemainingMs int64  `json:"remaining_ms,omitempty"`
}

type CompletionResponse struct {
	Success        bool   `json:"success"`
	DestinationURL string `json:"destination_url"`
}

type NFTBypassRequest struct {
	Message   string `json:"message" validate:"required"`
	Signature string `json:"signature" validate:"required"`
}
