package report

import (
	"strings"
)

// Generate formats the Dropbox changes into a readable report
func Generate(changes []string) string {
	if len(changes) == 0 {
		return "No changes detected in Dropbox today."
	}

	var sb strings.Builder
	sb.WriteString("📢 *Dropbox Daily Report*\n\n")
	sb.WriteString("Here are the changes detected in your Dropbox:\n\n")

	for _, change := range changes {
		sb.WriteString("- " + change + "\n")
	}

	sb.WriteString("\n📅 Report generated automatically.\n")

	return sb.String()
}

