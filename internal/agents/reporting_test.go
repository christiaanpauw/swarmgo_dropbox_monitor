package agents

import (
	"strings"
	"testing"
	"time"
)

func TestReportingAgent_Execute(t *testing.T) {
	// Create test data
	now := time.Now()
	oldTime := now.Add(-24 * time.Hour)

	testCases := []struct {
		name      string
		args      map[string]interface{}
		context   map[string]interface{}
		wantErr   bool
	}{
		{
			name: "Valid report generation",
			args: map[string]interface{}{
				"timeWindow": "24h",
				"format":     "html",
			},
			context: map[string]interface{}{
				"changes": []map[string]interface{}{
					{
						"Path":    "/test1.txt",
						"Type":    "modified",
						"ModTime": now,
						"Metadata": map[string]interface{}{
							"id":           "id1",
							"name":         "test1.txt",
							"path_display": "/test1.txt",
							"size":         100,
							"keywords":     []string{"test", "document"},
							"topics":       []string{"documentation"},
						},
					},
					{
						"Path":    "/test2.txt",
						"Type":    "modified",
						"ModTime": oldTime,
						"Metadata": map[string]interface{}{
							"id":           "id2",
							"name":         "test2.txt",
							"path_display": "/test2.txt",
							"size":         200,
							"keywords":     []string{"test", "code"},
							"topics":       []string{"source_code"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing timeWindow",
			args: map[string]interface{}{
				"format": "html",
			},
			context: map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "Invalid format",
			args: map[string]interface{}{
				"timeWindow": "24h",
				"format":     "invalid",
			},
			context: map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent := NewReportingAgent()
			result := agent.Execute(tc.args, tc.context)

			if tc.wantErr {
				if result.Success {
					t.Error("Expected error but got success")
				}
			} else {
				if !result.Success {
					t.Errorf("Expected success but got error: %v", result.Error)
				}

				// For valid cases, verify the report content
				if output, ok := result.Data.(string); ok {
					if tc.args["format"] == "html" {
						if !strings.Contains(output, "<!DOCTYPE html>") {
							t.Error("Expected HTML output")
						}
					} else {
						if strings.Contains(output, "<!DOCTYPE html>") {
							t.Error("Expected text output, got HTML")
						}
					}
				} else {
					t.Error("Expected string output")
				}
			}
		})
	}
}

func TestReportingAgent_GetTopItems(t *testing.T) {
	agent := NewReportingAgent()

	tests := []struct {
		name  string
		items map[string]int
		n     int
		want  int
	}{
		{
			name:  "Empty map",
			items: map[string]int{},
			n:     5,
			want:  0,
		},
		{
			name: "Less items than requested",
			items: map[string]int{
				"a": 1,
				"b": 2,
			},
			n:    5,
			want: 2,
		},
		{
			name: "More items than requested",
			items: map[string]int{
				"a": 1,
				"b": 2,
				"c": 3,
				"d": 4,
				"e": 5,
			},
			n:    3,
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agent.getTopItems(tt.items, tt.n)
			if len(got) != tt.want {
				t.Errorf("getTopItems() returned %d items, want %d", len(got), tt.want)
			}
		})
	}
}

func TestReportingAgent_GenerateReportSummary(t *testing.T) {
	agent := NewReportingAgent()

	now := time.Now()
	report := &Report{
		GeneratedAt: now,
		TimeWindow:  "24h",
		FileCount:   2,
		FilesByType: map[string]int{
			"document": 1,
			"source":   1,
		},
		TopKeywords: map[string]int{
			"test": 2,
			"code": 1,
		},
		TopTopics: map[string]int{
			"documentation": 1,
			"source_code":   1,
		},
		RecentFiles: []map[string]interface{}{
			{
				"Path":    "/test1.txt",
				"Type":    "modified",
				"ModTime": now,
			},
		},
	}

	summary := agent.generateReportSummary(report)

	// Verify summary content
	expectedParts := []string{
		"Report generated at",
		"Time window: 24h",
		"Total files: 2",
		"File types:",
		"document: 1",
		"source: 1",
		"Top keywords:",
		"test (2",
		"code (1",
		"Top topics:",
		"documentation (1",
		"source_code (1",
	}

	for _, part := range expectedParts {
		if !strings.Contains(summary, part) {
			t.Errorf("Summary missing expected part: %s", part)
		}
	}
}
