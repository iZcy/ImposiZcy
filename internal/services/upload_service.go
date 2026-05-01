package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/iZcy/imposizcy/config"
	"github.com/sirupsen/logrus"
)

type UploadService struct {
	cfg    *config.Config
	logger *logrus.Logger
}

func NewUploadService(cfg *config.Config, logger *logrus.Logger) *UploadService {
	return &UploadService{cfg: cfg, logger: logger}
}

// allowedExtensions defines permitted image file types
var allowedExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
	".svg":  true,
}

// SaveUpload saves an uploaded file to the upload directory.
// Returns the relative path (for DB storage) and the full filesystem path.
func (s *UploadService) SaveUpload(file *multipart.FileHeader, subDir string) (relativePath string, fullPath string, err error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		return "", "", fmt.Errorf("file type %s not allowed. permitted: png, jpg, jpeg, gif, bmp, webp, svg", ext)
	}

	// Check file size
	maxBytes := int64(s.cfg.Storage.MaxUploadMB) * 1024 * 1024
	if file.Size > maxBytes {
		return "", "", fmt.Errorf("file too large: %d bytes (max %d MB)", file.Size, s.cfg.Storage.MaxUploadMB)
	}

	// Create upload directory: uploads/<subDir>/
	destDir := filepath.Join(s.cfg.Storage.UploadDir, subDir)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	filename := fmt.Sprintf("%d%s", file.Size, ext)
	// Use original filename but make safe
	safeName := strings.ReplaceAll(file.Filename, " ", "_")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	safeName = strings.ReplaceAll(safeName, "\\", "_")
	if safeName != "" {
		filename = safeName
	}

	fullPath = filepath.Join(destDir, filename)
	relativePath = filepath.Join(subDir, filename)

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(fullPath)
		return "", "", fmt.Errorf("failed to save file: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"filename": filename,
		"subDir":   subDir,
		"size":     file.Size,
	}).Info("File uploaded successfully")

	return relativePath, fullPath, nil
}

// DeleteUpload removes an uploaded file
func (s *UploadService) DeleteUpload(relativePath string) error {
	fullPath := filepath.Join(s.cfg.Storage.UploadDir, relativePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// EnsureUploadDir creates the upload directory if it doesn't exist
func (s *UploadService) EnsureUploadDir() error {
	return os.MkdirAll(s.cfg.Storage.UploadDir, 0755)
}
