package analysis

import (
	"context"
	"crypto/sha256"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/christiaanpauw/swarmgo_dropbox_monitor/internal/models"
)

// ContentAnalyzer analyzes file content
type ContentAnalyzer interface {
	AnalyzeContent(ctx context.Context, path string, content []byte) (*models.FileContent, error)
}

// contentAnalyzer implements the ContentAnalyzer interface
type contentAnalyzer struct{}

// NewContentAnalyzer creates a new content analyzer
func NewContentAnalyzer() ContentAnalyzer {
	return &contentAnalyzer{}
}

// AnalyzeContent analyzes the content of a file and returns metadata about it
func (a *contentAnalyzer) AnalyzeContent(ctx context.Context, path string, content []byte) (*models.FileContent, error) {
	// Get file extension and MIME type
	ext := filepath.Ext(path)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = http.DetectContentType(content)
	}

	// Create file content analysis
	analysis := &models.FileContent{
		Path:         path,
		ContentType:  mimeType,
		Size:         int64(len(content)),
		IsBinary:     !isTextFile(content),
		ContentHash:  calculateHash(content),
	}

	return analysis, nil
}

// isTextFile checks if the content appears to be text
func isTextFile(content []byte) bool {
	if len(content) == 0 {
		return true
	}

	// Check for null bytes which typically indicate binary content
	for _, b := range content {
		if b == 0 {
			return false
		}
	}

	return true
}

// calculateHash generates a hash of the content
func calculateHash(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return fmt.Sprintf("%x", h.Sum(nil))
}
