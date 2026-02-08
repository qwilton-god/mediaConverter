package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"mediaConverter/api/dto"
	"mediaConverter/api/middleware"
	"mediaConverter/api/service"
)

type TaskHandler struct {
	service *service.TaskService
	logger  *zap.Logger
}

func NewTaskHandler(service *service.TaskService, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{
		service: service,
		logger:  logger,
	}
}

func (h *TaskHandler) Upload(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.GetTraceID(r.Context())

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.handleError(w, "Failed to parse form", err, traceID, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.handleError(w, "Failed to get file", err, traceID, http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := h.validateFile(header, file); err != nil {
		h.handleError(w, "Invalid file", err, traceID, http.StatusBadRequest)
		return
	}

	filename := sanitizeFilename(header.Filename)
	filePath := filepath.Join("/uploads", filename)

	dst, err := os.Create(filePath)
	if err != nil {
		h.handleError(w, "Failed to save file", err, traceID, http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		h.handleError(w, "Failed to write file", err, traceID, http.StatusInternalServerError)
		return
	}

	req := &dto.CreateTaskRequest{
		OriginalFilename: header.Filename,
		FilePath:         filePath,
	}

	resp, err := h.service.CreateTask(r.Context(), traceID, req)
	if err != nil {
		h.handleError(w, "Failed to create task", err, traceID, http.StatusInternalServerError)
		return
	}

	h.logger.Info("File uploaded",
		zap.String("trace_id", traceID),
		zap.String("task_id", resp.ID),
		zap.String("filename", header.Filename),
	)

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *TaskHandler) Status(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.GetTraceID(r.Context())

	taskID := strings.TrimPrefix(r.URL.Path, "/status/")
	if taskID == "" {
		h.handleError(w, "Task ID is required", nil, traceID, http.StatusBadRequest)
		return
	}

	resp, err := h.service.GetTaskStatus(r.Context(), taskID)
	if err != nil {
		if errors.Is(err, dto.ErrTaskNotFound) {
			h.handleError(w, "Task not found", err, traceID, http.StatusNotFound)
			return
		}
		h.handleError(w, "Failed to get task status", err, traceID, http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *TaskHandler) validateFile(header *multipart.FileHeader, file multipart.File) error {
	const maxSize = 100 * 1024 * 1024

	if header.Size > maxSize {
		return errors.New("file too large")
	}

	if !isAllowedFileType(header.Filename) {
		return errors.New("invalid file type")
	}

	return nil
}

func isAllowedFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
		".mp4":  true,
	}
	return allowed[ext]
}

func sanitizeFilename(filename string) string {
	return filepath.Base(filename)
}

func (h *TaskHandler) handleError(w http.ResponseWriter, message string, err error, traceID string, status int) {
	h.logger.Error(message,
		zap.String("trace_id", traceID),
		zap.Error(err),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.ErrorResponse{
		Error:   message,
		TraceID: traceID,
	})
}

func (h *TaskHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
