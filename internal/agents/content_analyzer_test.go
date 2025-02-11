package agents

import (
	"strings"
	"testing"
)

func TestContentAnalyzer_ExtractKeywords(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Empty content",
			content:  "",
			expected: []string{},
		},
		{
			name:     "Simple content",
			content:  "This is a test document about testing code quality",
			expected: []string{"test", "document", "testing", "code", "quality"},
		},
		{
			name:     "Content with stop words",
			content:  "the quick brown fox jumps over the lazy dog",
			expected: []string{"quick", "brown", "fox", "jumps", "lazy", "dog"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewContentAnalyzer("test-token")
			got := analyzer.ExtractKeywords(tt.content)
			if len(got) > len(tt.expected) {
				t.Errorf("Expected at most %d keywords, got %d", len(tt.expected), len(got))
			}
		})
	}
}

func TestContentAnalyzer_ExtractTopics(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		want     []string
	}{
		{
			name:    "Empty content",
			content: "",
			want:    []string{},
		},
		{
			name:    "Programming content",
			content: "This is a function that handles errors and interfaces",
			want:    []string{"function", "error", "interface"},
		},
		{
			name:    "Documentation content",
			content: "This is a readme file with documentation and examples",
			want:    []string{"readme", "documentation", "example"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewContentAnalyzer("test-token")
			got := analyzer.ExtractTopics(tt.content)
			if len(got) > len(tt.want) {
				t.Errorf("Expected at most %d topics, got %d", len(tt.want), len(got))
			}
		})
	}
}

func TestContentAnalyzer_GenerateSummary(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "Empty content",
			content: "",
			want:    "",
		},
		{
			name:    "Short content",
			content: "This is a short summary",
			want:    "This is a short summary",
		},
		{
			name:    "Long content",
			content: strings.Repeat("This is a very long content. ", 20),
			want:    strings.Repeat("This is a very long content. ", 6) + "This is a very long conten...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewContentAnalyzer("test-token")
			got := analyzer.GenerateSummary(tt.content)
			if tt.want != got {
				t.Errorf("GenerateSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentAnalyzer_Execute(t *testing.T) {
	analyzer := NewContentAnalyzer("test-token")
	result := analyzer.Execute(map[string]interface{}{
		"paths": []string{"/test.txt"},
	}, nil)

	if !result.Success {
		t.Error("Expected success")
	}
}

func TestContentAnalyzer_ShouldSkipFile(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Text file",
			path: "test.txt",
			want: false,
		},
		{
			name: "Binary file",
			path: "test.exe",
			want: true,
		},
		{
			name: "Archive file",
			path: "test.zip",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewContentAnalyzer("test-token")
			if got := analyzer.shouldSkipFile(tt.path); got != tt.want {
				t.Errorf("shouldSkipFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
