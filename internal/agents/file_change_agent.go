package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/prathyushnallamothu/swarmgo"
)

type fileChangeAgent struct {
	dbxConfig dropbox.Config
	client    DropboxFilesClient
}

func NewFileChangeAgent(token string) FileChangeAgent {
	config := dropbox.Config{
		Token: token,
	}
	return &fileChangeAgent{
		dbxConfig: config,
		client:    files.New(config),
	}
}

func (fca *fileChangeAgent) DetectChanges(ctx context.Context, timeWindow string) ([]FileChange, error) {
	// Parse time window
	duration, err := time.ParseDuration(timeWindow)
	if err != nil {
		return nil, fmt.Errorf("invalid time window: %v", err)
	}

	// Calculate start time
	startTime := time.Now().Add(-duration)

	// List folder
	arg := files.NewListFolderArg("")
	arg.Recursive = true
	result, err := fca.client.ListFolder(arg)
	if err != nil {
		return nil, fmt.Errorf("error listing folder: %v", err)
	}

	var changes []FileChange
	for _, entry := range result.Entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			if f.ServerModified.After(startTime) {
				changes = append(changes, FileChange{
					Path:    f.PathLower,
					ModTime: f.ServerModified.Format(time.RFC3339),
					Metadata: map[string]interface{}{
						"id":            f.Id,
						"name":          f.Name,
						"size":          f.Size,
						"content_hash":  f.ContentHash,
						"path_display":  f.PathDisplay,
						"is_downloadable": f.IsDownloadable,
					},
				})
			}
		}
	}

	// Continue listing if there are more entries
	for result.HasMore {
		arg := files.NewListFolderContinueArg(result.Cursor)
		result, err = fca.client.ListFolderContinue(arg)
		if err != nil {
			return nil, fmt.Errorf("error continuing folder listing: %v", err)
		}

		for _, entry := range result.Entries {
			switch f := entry.(type) {
			case *files.FileMetadata:
				if f.ServerModified.After(startTime) {
					changes = append(changes, FileChange{
						Path:    f.PathLower,
						ModTime: f.ServerModified.Format(time.RFC3339),
						Metadata: map[string]interface{}{
							"id":            f.Id,
							"name":          f.Name,
							"size":          f.Size,
							"content_hash":  f.ContentHash,
							"path_display":  f.PathDisplay,
							"is_downloadable": f.IsDownloadable,
						},
					})
				}
			}
		}
	}

	return changes, nil
}

func (fca *fileChangeAgent) Execute(args map[string]interface{}, contextVariables map[string]interface{}) swarmgo.Result {
	timeWindow := args["timeWindow"].(string)

	// Detect changes
	changes, err := fca.DetectChanges(context.Background(), timeWindow)
	if err != nil {
		return swarmgo.Result{
			Success: false,
			Error:   err,
		}
	}

	return swarmgo.Result{
		Success: true,
		Data:    changes,
	}
}
