package analysis

import (
	"fmt"
	"hallucino/internal/k8s"
	"regexp"
)

// LogAnalyzer provides methods for processing Kubernetes logs
type LogAnalyzer struct {
	logs              []k8s.LogEntry
	criticalEvents    []k8s.LogEntry
	performanceIssues []k8s.LogEntry
	errorCount        int
	warningCount      int
}

// NewLogAnalyzer creates a new log analyzer instance
func NewLogAnalyzer(logs []k8s.LogEntry) *LogAnalyzer {
	la := &LogAnalyzer{
		logs:              logs,
		errorCount:        0,
		warningCount:      0,
		criticalEvents:    []k8s.LogEntry{},
		performanceIssues: []k8s.LogEntry{},
	}
	la.processLogs()
	return la
}

// processLogs analyzes all log entries
func (la *LogAnalyzer) processLogs() {
	for _, log := range la.logs {
		la.analyzeLine(log)
	}
}

// analyzeLine performs detailed analysis of each log line
func (la *LogAnalyzer) analyzeLine(log k8s.LogEntry) {
	errorRegex := regexp.MustCompile(`(?i)error|critical|fatal|panic`)
	warningRegex := regexp.MustCompile(`(?i)warning|warn`)
	performanceRegex := regexp.MustCompile(`(?i)timeout|latency|slow|high load`)
	restartRegex := regexp.MustCompile(`(?i)pod|container.*restart`)

	switch {
	case errorRegex.MatchString(log.LogContent):
		la.errorCount++
		la.criticalEvents = append(la.criticalEvents, log)
	case warningRegex.MatchString(log.LogContent):
		la.warningCount++
	case performanceRegex.MatchString(log.LogContent):
		la.performanceIssues = append(la.performanceIssues, log)
	case restartRegex.MatchString(log.LogContent):
		log.LogContent = "Restart Event: " + log.LogContent
		la.criticalEvents = append(la.criticalEvents, log)
	}
}

// generateDetailedReport creates a comprehensive log analysis report
func (la *LogAnalyzer) generateDetailedReport() string {
	report := "### Kubernetes Log Analysis Report\n\n"
	report += fmt.Sprintf("- **Total Log Entries:** %d\n", len(la.logs))
	report += fmt.Sprintf("- **Error Count:** %d\n", la.errorCount)
	report += fmt.Sprintf("- **Warning Count:** %d\n\n", la.warningCount)

	report += "#### Critical Events\n"
	if len(la.criticalEvents) > 0 {
		for _, event := range la.criticalEvents {
			report += fmt.Sprintf("- `%s | %s | %s`: %s\n",
				event.Timestamp,
				event.PodName,
				event.Container,
				event.LogContent,
			)
		}
	} else {
		report += "- No critical events detected.\n"
	}

	report += "\n#### Performance Issues\n"
	if len(la.performanceIssues) > 0 {
		for _, issue := range la.performanceIssues {
			report += fmt.Sprintf("- `%s | %s | %s`: %s\n",
				issue.Timestamp,
				issue.PodName,
				issue.Container,
				issue.LogContent,
			)
		}
	} else {
		report += "- No significant performance issues detected.\n"
	}

	return report
}
