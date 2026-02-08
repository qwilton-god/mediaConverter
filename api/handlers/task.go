package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
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
	"mediaConverter/api/validation"
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

	outputFormat := r.FormValue("output_format")
	var targetWidth, targetHeight *int
	if w := r.FormValue("target_width"); w != "" {
		width := 0
		if _, err := fmt.Sscanf(w, "%d", &width); err == nil {
			targetWidth = &width
		}
	}
	if h := r.FormValue("target_height"); h != "" {
		height := 0
		if _, err := fmt.Sscanf(h, "%d", &height); err == nil {
			targetHeight = &height
		}
	}
	crop := r.FormValue("crop") == "true"

	req := &dto.CreateTaskRequest{
		OriginalFilename: header.Filename,
		FilePath:         filePath,
		OutputFormat:     outputFormat,
		TargetWidth:      targetWidth,
		TargetHeight:     targetHeight,
		Crop:             crop,
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
		h.logger.Warn("File too large",
			zap.String("filename", header.Filename),
			zap.Int64("size", header.Size),
		)
		return validation.ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))

	fileType, err := validation.DetectFileType(file)
	if err != nil {
		h.logger.Warn("Magic bytes detection failed",
			zap.String("filename", header.Filename),
			zap.Error(err),
		)
		return validation.ErrInvalidFileType
	}

	extToType := map[string]validation.FileType{
		".jpg":  validation.FileTypeJPEG,
		".jpeg": validation.FileTypeJPEG,
		".png":  validation.FileTypePNG,
		".gif":  validation.FileTypeGIF,
		".pdf":  validation.FileTypePDF,
	}

	expectedType, ok := extToType[ext]
	if !ok {
		h.logger.Warn("Unsupported file extension",
			zap.String("filename", header.Filename),
			zap.String("extension", ext),
		)
		return validation.ErrUnsupportedFormat
	}

	if fileType != expectedType {
		h.logger.Warn("File extension mismatch with magic bytes",
			zap.String("filename", header.Filename),
			zap.String("extension", ext),
			zap.String("expected_type", string(expectedType)),
			zap.String("detected_type", string(fileType)),
		)
		return validation.ErrExtensionMismatch
	}

	return nil
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
