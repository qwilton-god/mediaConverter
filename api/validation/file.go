package validation

import (
	"bytes"
	"io"
	"mime/multipart"
)

type FileType string

const (
	FileTypePNG  FileType = "png"
	FileTypeJPEG FileType = "jpeg"
	FileTypeGIF  FileType = "gif"
	FileTypePDF  FileType = "pdf"
	FileTypeMP4  FileType = "mp4"
)

var magicBytes = map[FileType][]byte{
	FileTypePNG:  {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
	FileTypeJPEG: {0xFF, 0xD8, 0xFF},
	FileTypeGIF:  {0x47, 0x49, 0x46, 0x38},
	FileTypePDF:  {0x25, 0x50, 0x44, 0x46},
}

func DetectFileType(file multipart.File) (FileType, error) {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}

	for fileType, signature := range magicBytes {
		if bytes.HasPrefix(buffer[:n], signature) {
			return fileType, nil
		}
	}

	return "", ErrInvalidFileType
}

func IsAllowedImageType(fileType FileType) bool {
	switch fileType {
	case FileTypePNG, FileTypeJPEG, FileTypeGIF:
		return true
	default:
		return false
	}
}
