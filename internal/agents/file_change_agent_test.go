package agents

import (
	"errors"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

// Mock Dropbox client
type mockDropboxClient struct {
	entries []files.IsMetadata
	hasMore bool
	cursor  string
	err     error
}

func (m *mockDropboxClient) ListFolder(arg *files.ListFolderArg) (res *files.ListFolderResult, err error) {
	if m.err != nil {
		return nil, m.err
	}
	return &files.ListFolderResult{
		Entries:  m.entries,
		HasMore:  m.hasMore,
		Cursor:   m.cursor,
	}, nil
}

func (m *mockDropboxClient) ListFolderContinue(arg *files.ListFolderContinueArg) (res *files.ListFolderResult, err error) {
	if m.err != nil {
		return nil, m.err
	}
	return &files.ListFolderResult{
		Entries:  m.entries,
		HasMore:  false,
		Cursor:   "",
	}, nil
}

func TestFileChangeAgent_Execute(t *testing.T) {
	// Create test data
	now := time.Now()
	oldTime := now.Add(-24 * time.Hour)
	
	testCases := []struct {
		name      string
		entries   []files.IsMetadata
		hasMore   bool
		cursor    string
		err       error
		timeWindow string
		wantErr   bool
		wantCount int
	}{
		{
			name: "Recent changes found",
			entries: []files.IsMetadata{
				&files.FileMetadata{
					Metadata: files.Metadata{
						Name:        "test1.txt",
						PathDisplay: "/test1.txt",
					},
					ServerModified: now,
					Id:            "id1",
					Size:          100,
					ContentHash:   "hash1",
				},
				&files.FileMetadata{
					Metadata: files.Metadata{
						Name:        "test2.txt",
						PathDisplay: "/test2.txt",
					},
					ServerModified: oldTime,
					Id:            "id2",
					Size:          200,
					ContentHash:   "hash2",
				},
			},
			hasMore:    false,
			cursor:     "",
			err:        nil,
			timeWindow: "12h",
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:       "Invalid time window",
			entries:    nil,
			hasMore:    false,
			cursor:     "",
			err:        nil,
			timeWindow: "invalid",
			wantErr:    true,
			wantCount:  0,
		},
		{
			name:       "Dropbox error",
			entries:    nil,
			hasMore:    false,
			cursor:     "",
			err:        errors.New("dropbox error"),
			timeWindow: "12h",
			wantErr:    true,
			wantCount:  0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create agent with mock client
			agent := NewFileChangeAgent("test-token")
			agent.client = &mockDropboxClient{
				entries: tc.entries,
				hasMore: tc.hasMore,
				cursor: tc.cursor,
				err:    tc.err,
			}

			// Execute agent
			args := map[string]interface{}{
				"timeWindow": tc.timeWindow,
			}
			contextVars := map[string]interface{}{}

			result := agent.Execute(args, contextVars)

			if tc.wantErr {
				if result.Success {
					t.Error("Expected error but got success")
				}
			} else {
				if !result.Success {
					t.Errorf("Expected success but got error: %v", result.Error)
				}

				changes, ok := result.Data.([]map[string]interface{})
				if !ok {
					t.Error("Expected []map[string]interface{} result")
					return
				}

				if len(changes) != tc.wantCount {
					t.Errorf("Expected %d changes, got %d", tc.wantCount, len(changes))
				}

				// For successful cases with changes, verify the first change
				if len(changes) > 0 {
					change := changes[0]
					if path, ok := change["Path"].(string); !ok || path != "/test1.txt" {
						t.Errorf("Expected path /test1.txt, got %v", path)
					}
					if changeType, ok := change["Type"].(string); !ok || changeType != "modified" {
						t.Errorf("Expected type 'modified', got %v", changeType)
					}
				}
			}
		})
	}
}

func TestFileChangeAgent_GetChanges(t *testing.T) {
	now := time.Now()
	
	testCases := []struct {
		name     string
		duration time.Duration
		entries  []files.IsMetadata
		hasMore  bool
		cursor   string
		err      error
		want     int
		wantErr  bool
	}{
		{
			name:     "Recent files",
			duration: 12 * time.Hour,
			entries: []files.IsMetadata{
				&files.FileMetadata{
					Metadata: files.Metadata{
						Name:        "test1.txt",
						PathDisplay: "/test1.txt",
					},
					ServerModified: now,
				},
				&files.FileMetadata{
					Metadata: files.Metadata{
						Name:        "test2.txt",
						PathDisplay: "/test2.txt",
					},
					ServerModified: now.Add(-24 * time.Hour),
				},
			},
			hasMore:  false,
			cursor:   "",
			err:      nil,
			want:     1,
			wantErr:  false,
		},
		{
			name:     "Dropbox error",
			duration: 12 * time.Hour,
			entries:  nil,
			hasMore:  false,
			cursor:   "",
			err:      errors.New("dropbox error"),
			want:     0,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent := NewFileChangeAgent("test-token")
			agent.client = &mockDropboxClient{
				entries: tc.entries,
				hasMore: tc.hasMore,
				cursor: tc.cursor,
				err:    tc.err,
			}

			changes, err := agent.getChanges(tc.duration)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if len(changes) != tc.want {
					t.Errorf("Expected %d changes, got %d", tc.want, len(changes))
				}
			}
		})
	}
}
