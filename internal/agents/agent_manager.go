package agents

import (
	"context"
	"fmt"
	"log"
)

// Agent interface as required by swarmgo
type Agent interface {
	Execute(ctx context.Context, input interface{}) (interface{}, error)
}

// AgentManager manages all agents in the system
type AgentManager struct {
	fileChange     FileChangeAgent
	database       DatabaseAgent
	contentAnalyzer ContentAnalyzer
	reporting      ReportingAgent
}

// NewAgentManager creates a new agent manager
func NewAgentManager(dbConnStr string, dropboxToken string) (*AgentManager, error) {
	// Create agents
	fileChange := NewFileChangeAgent(dropboxToken)
	database := NewDatabaseAgent(dbConnStr)
	contentAnalyzer := NewContentAnalyzer(dropboxToken)
	reporting := NewReportingAgent()

	return &AgentManager{
		fileChange:     fileChange,
		database:       database,
		contentAnalyzer: contentAnalyzer,
		reporting:      reporting,
	}, nil
}

// ProcessChanges orchestrates the entire workflow
func (am *AgentManager) ProcessChanges(ctx context.Context, timeWindow string) (*Report, error) {
	// Step 1: Identify changed files
	changes, err := am.fileChange.DetectChanges(ctx, timeWindow)
	if err != nil {
		return nil, fmt.Errorf("error detecting file changes: %v", err)
	}

	// Step 2: Store changes in database
	for _, change := range changes {
		err := am.database.StoreChange(ctx, change)
		if err != nil {
			log.Printf("Error storing change for %s: %v", change.Path, err)
			continue
		}

		// Step 3: Analyze content with Google AI Studio
		content, err := am.contentAnalyzer.AnalyzeFile(ctx, change.Path)
		if err != nil {
			log.Printf("Error analyzing content for %s: %v", change.Path, err)
			continue
		}
		
		// Store analysis results
		err = am.database.StoreAnalysis(ctx, change.Path, content)
		if err != nil {
			log.Printf("Error storing analysis for %s: %v", change.Path, err)
			continue
		}
	}

	// Step 4: Generate report including content analysis
	report, err := am.reporting.GenerateReport(ctx, changes)
	if err != nil {
		return nil, fmt.Errorf("error generating report: %v", err)
	}

	return report, nil
}
