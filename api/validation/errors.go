package validation

import "errors"

var (
	ErrInvalidFileType   = errors.New("invalid file type")
	ErrFileTooLarge      = errors.New("file size exceeds 100MB limit")
	ErrExtensionMismatch = errors.New("file extension does not match content")
	ErrUnsupportedFormat = errors.New("unsupported file format")
)
