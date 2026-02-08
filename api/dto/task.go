package dto

import "errors"

var ErrTaskNotFound = errors.New("task not found")

type CreateTaskRequest struct {
	OriginalFilename string `json:"original_filename"`
	FilePath         string `json:"file_path"`
}

type TaskResponse struct {
	ID               string  `json:"id"`
	TraceID          string  `json:"trace_id"`
	OriginalFilename string  `json:"original_filename"`
	Status           string  `json:"status"`
	ErrorMessage     string  `json:"error_message,omitempty"`
	CreatedAt        string  `json:"created_at"`
	CompletedAt      *string `json:"completed_at,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}
