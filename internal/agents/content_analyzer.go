package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type contentAnalyzer struct {
	dropboxToken string
	apiKey      string
}

func NewContentAnalyzer(dropboxToken string) ContentAnalyzer {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
	}

	apiKey := os.Getenv("GOOGLE_AISTUDIO_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: GOOGLE_AISTUDIO_API_KEY not set")
	}

	return &contentAnalyzer{
		dropboxToken: dropboxToken,
		apiKey:      apiKey,
	}
}

func (ca *contentAnalyzer) AnalyzeFile(ctx context.Context, path string) (*FileContent, error) {
	// First download file content from Dropbox
	content, err := ca.downloadFromDropbox(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("error downloading file: %v", err)
	}

	// Analyze content using Google AI Studio API
	analysis, err := ca.analyzeWithGoogleAI(ctx, string(content))
	if err != nil {
		return nil, fmt.Errorf("error analyzing content: %v", err)
	}

	return &FileContent{
		Path:        path,
		Summary:     analysis.Summary,
		Keywords:    analysis.Keywords,
		Categories:  analysis.Categories,
		Sensitivity: analysis.Sensitivity,
	}, nil
}

func (ca *contentAnalyzer) analyzeWithGoogleAI(ctx context.Context, content string) (*AnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze the following content and provide:
1. A brief summary
2. Key topics/keywords (comma-separated)
3. Categories (comma-separated)
4. Sensitivity level (Public, Internal, Confidential)

Content to analyze:
%s`, content)

	var reqBody struct {
		Contents struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"contents"`
	}
	
	reqBody.Contents.Parts = []struct {
		Text string `json:"text"`
	}{{Text: prompt}}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", 
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	q := req.URL.Query()
	q.Add("key", ca.apiKey)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var aiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if len(aiResp.Candidates) == 0 || len(aiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	// Parse the response
	analysis := &AnalysisResult{}
	lines := bytes.Split([]byte(aiResp.Candidates[0].Content.Parts[0].Text), []byte("\n"))
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("1.")) {
			analysis.Summary = string(bytes.TrimSpace(bytes.TrimPrefix(line, []byte("1."))))
		} else if bytes.HasPrefix(line, []byte("2.")) {
			keywords := bytes.Split(bytes.TrimSpace(bytes.TrimPrefix(line, []byte("2."))), []byte(","))
			for _, k := range keywords {
				analysis.Keywords = append(analysis.Keywords, string(bytes.TrimSpace(k)))
			}
		} else if bytes.HasPrefix(line, []byte("3.")) {
			categories := bytes.Split(bytes.TrimSpace(bytes.TrimPrefix(line, []byte("3."))), []byte(","))
			for _, c := range categories {
				analysis.Categories = append(analysis.Categories, string(bytes.TrimSpace(c)))
			}
		} else if bytes.HasPrefix(line, []byte("4.")) {
			analysis.Sensitivity = string(bytes.TrimSpace(bytes.TrimPrefix(line, []byte("4."))))
		}
	}

	return analysis, nil
}

func (ca *contentAnalyzer) downloadFromDropbox(ctx context.Context, path string) ([]byte, error) {
	// Implementation of Dropbox download
	// This would use the dropbox-sdk-go library to download the file
	// For now returning empty implementation
	return []byte(""), nil
}
