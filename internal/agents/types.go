package agents

import (
	"context"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

// DropboxFilesClient is a minimal interface for the Dropbox files client
type DropboxFilesClient interface {
	ListFolder(arg *files.ListFolderArg) (*files.ListFolderResult, error)
	ListFolderContinue(arg *files.ListFolderContinueArg) (*files.ListFolderResult, error)
}

// FileChange represents a change in a file
type FileChange struct {
	Path     string
	ModTime  string
	Metadata map[string]interface{}
}

// FileContent represents analyzed file content
type FileContent struct {
	Path        string
	Summary     string
	Keywords    []string
	Categories  []string
	Sensitivity string
}

// Report represents the final report
type Report struct {
	GeneratedAt  time.Time
	TimeWindow   string
	FileCount    int
	FilesByType  map[string]int
	TopKeywords  map[string]int
	TopTopics    map[string]int
	RecentFiles  []map[string]interface{}
	Summary      string
	Changes      []FileChange
}

// GoogleAIRequest represents a request to the Google AI API
type GoogleAIRequest struct {
	Contents struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

// GoogleAIResponse represents a response from the Google AI API
type GoogleAIResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// AnalysisResult represents the result of content analysis
type AnalysisResult struct {
	Summary     string
	Keywords    []string
	Categories  []string
	Sensitivity string
}

// Agent interfaces
type (
	// FileChangeAgent detects changes in files
	FileChangeAgent interface {
		DetectChanges(ctx context.Context, timeWindow string) ([]FileChange, error)
	}

	// DatabaseAgent handles database operations
	DatabaseAgent interface {
		StoreChange(ctx context.Context, change FileChange) error
		StoreAnalysis(ctx context.Context, path string, content *FileContent) error
		Close() error
	}

	// ContentAnalyzer analyzes file content
	ContentAnalyzer interface {
		AnalyzeFile(ctx context.Context, path string) (*FileContent, error)
	}

	// ReportingAgent generates reports
	ReportingAgent interface {
		GenerateReport(ctx context.Context, changes []FileChange) (*Report, error)
	}
)
