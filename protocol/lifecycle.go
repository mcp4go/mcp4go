package protocol

import "encoding/json"

// CancelRequest is sent to cancel a pending request
type CancelRequest struct {
	// ID of the request to cancel
	ID json.RawMessage `json:"id"`
}

// CancelResult confirms the cancellation
type CancelResult struct{}

// ProgressToken identifies a specific ongoing operation
type ProgressToken string

// ProgressNotification reports progress for long-running operations
type ProgressNotification struct {
	// Token identifies the operation
	Token ProgressToken `json:"token"`
	// Value indicates the progress (0-100%)
	Value int `json:"value"`
	// Message provides additional progress information
	Message string `json:"message,omitempty"`
}
