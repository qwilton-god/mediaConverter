package dto

import "errors"

var ErrTaskNotFound = errors.New("task not found")

type CreateTaskRequest struct {
	OriginalFilename string `json:"original_filename"`
	FilePath         string `json:"file_path"`
	OutputFormat     string `json:"output_format"`
	TargetWidth      *int   `json:"target_width"`
	TargetHeight     *int   `json:"target_height"`
	Crop             bool   `json:"crop"`
}

type TaskResponse struct {
	ID                string  `json:"id"`
	TraceID           string  `json:"trace_id"`
	OriginalFilename  string  `json:"original_filename"`
	OutputFilename    string  `json:"output_filename,omitempty"`
	OutputFormat      string  `json:"output_format"`
	TargetWidth       *int    `json:"target_width,omitempty"`
	TargetHeight      *int    `json:"target_height,omitempty"`
	Crop              bool    `json:"crop"`
	Status            string  `json:"status"`
	ErrorMessage      string  `json:"error_message,omitempty"`
	CreatedAt         string  `json:"created_at"`
	CompletedAt       *string `json:"completed_at,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}
