package uixt

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// DetectAndRenameImageFile examines the file content to determine its media type
// and renames the file with the appropriate extension (.jpg, .png, .mp4, etc.)
func DetectAndRenameMediaFile(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for type detection: %v", err)
	}
	defer file.Close()

	// Read the first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for type detection: %v", err)
	}

	// Reset file pointer
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %v", err)
	}

	// Detect content type
	contentType := http.DetectContentType(buffer)
	log.Info().Str("filePath", filePath).Str("contentType", contentType).Msg("Detected content type")

	// Determine file extension based on content type
	var extension string
	switch {
	// Image types
	case strings.Contains(contentType, "image/jpeg"):
		extension = ".jpg"
	case strings.Contains(contentType, "image/png"):
		extension = ".png"
	case strings.Contains(contentType, "image/gif"):
		extension = ".gif"
	case strings.Contains(contentType, "image/webp"):
		extension = ".webp"
	case strings.Contains(contentType, "image/bmp"):
		extension = ".bmp"
	case strings.Contains(contentType, "image/tiff"):
		extension = ".tiff"
	case strings.Contains(contentType, "image/svg+xml"):
		extension = ".svg"

	// Video types
	case strings.Contains(contentType, "video/mp4"):
		extension = ".mp4"
	case strings.Contains(contentType, "video/quicktime"):
		extension = ".mov"
	case strings.Contains(contentType, "video/x-msvideo"):
		extension = ".avi"
	case strings.Contains(contentType, "video/x-ms-wmv"):
		extension = ".wmv"
	case strings.Contains(contentType, "video/x-flv"):
		extension = ".flv"
	case strings.Contains(contentType, "video/webm"):
		extension = ".webm"
	case strings.Contains(contentType, "video/x-matroska"):
		extension = ".mkv"

	default:
		// Check for general image or video types
		if strings.Contains(contentType, "image/") {
			extension = ".jpg" // Default for unknown image types
		} else if strings.Contains(contentType, "video/") {
			extension = ".mp4" // Default for unknown video types
		} else {
			// Try to determine from original file extension
			origExt := strings.ToLower(filepath.Ext(filePath))
			if origExt == ".mp4" || origExt == ".mov" || origExt == ".avi" ||
				origExt == ".wmv" || origExt == ".flv" || origExt == ".webm" || origExt == ".mkv" {
				extension = origExt
			} else if origExt == ".jpg" || origExt == ".jpeg" || origExt == ".png" ||
				origExt == ".gif" || origExt == ".webp" || origExt == ".bmp" ||
				origExt == ".tiff" || origExt == ".svg" {
				extension = origExt
			} else {
				return filePath, fmt.Errorf("not a recognized media type: %s", contentType)
			}
		}
	}

	// Create new file path with extension
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	newFilePath := filepath.Join(dir, base+extension)

	// If the file already has the correct extension, just return it
	if filePath == newFilePath {
		return filePath, nil
	}

	// Rename the file
	err = os.Rename(filePath, newFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to rename file: %v", err)
	}

	log.Info().Str("oldPath", filePath).Str("newPath", newFilePath).Msg("Renamed image file with proper extension")
	return newFilePath, nil
}
