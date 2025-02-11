package analysis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContentAnalyzer_AnalyzeContent(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		content     []byte
		wantBinary  bool
		wantMIME    string
		wantErrNil  bool
	}{
		{
			name:        "Text file analysis",
			path:        "test.txt",
			content:     []byte("Hello, World!"),
			wantBinary:  false,
			wantMIME:    "text/plain; charset=utf-8",
			wantErrNil:  true,
		},
		{
			name:        "Binary file analysis",
			path:        "test.bin",
			content:     []byte{0x00, 0x01, 0x02, 0x03},
			wantBinary:  true,
			wantMIME:    "application/octet-stream",
			wantErrNil:  true,
		},
		{
			name:        "Empty file analysis",
			path:        "empty.txt",
			content:     []byte{},
			wantBinary:  false,
			wantMIME:    "text/plain; charset=utf-8",
			wantErrNil:  true,
		},
	}

	analyzer := NewContentAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeContent(context.Background(), tt.path, tt.content)

			if tt.wantErrNil {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.path, result.Path)
				assert.Equal(t, tt.wantBinary, result.IsBinary)
				assert.Equal(t, tt.wantMIME, result.ContentType)
				assert.Equal(t, int64(len(tt.content)), result.Size)
				assert.NotEmpty(t, result.ContentHash)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
