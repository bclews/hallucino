package analysis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

// AnalysisPrompt is the constant for guiding log analysis
const AnalysisPrompt = `**Instructions:**
You are an expert in analyzing Kubernetes logs. Your goal is to analyze the given log data, identify patterns, detect anomalies, and summarize insights clearly. Pay attention to the context provided in the logs, and focus on:
1. **Errors and Warnings:** Highlight any critical issues, error messages, or warnings that may indicate failures or potential problems in the Kubernetes system or its components.
2. **Patterns and Trends:** Look for repetitive log entries, trends, or sequences that might indicate systemic issues or misconfigurations.
3. **System Events:** Summarize key system events such as service startups, shutdowns, or resource state changes.
4. **Performance Issues:** Identify potential bottlenecks, timeouts, or latency-related concerns.
5. **Suggestions:** Provide actionable recommendations for resolving any detected issues or improving the system's stability and performance.

**Example Format of Response:**
* **Summary of Key Events:** (List of major activities or noteworthy occurrences.)
* **Detected Issues and Errors:** (Detailed analysis of errors, anomalies, or potential issues.)
* **Pattern Observations:** (Summary of any recurring patterns or trends in the logs.)
* **Actionable Recommendations:** (Specific steps or insights to address the issues identified.)`

// Config represents the configuration for OpenAI
type Config struct {
	APIKey         string
	Endpoint       string
	DeploymentName string
}

// OpenAIAnalyzer handles AI-powered log insights generation
type OpenAIAnalyzer struct {
	client *azopenai.Client
	config Config
}

// NewOpenAIAnalyzer creates a new OpenAI log analyzer
func NewOpenAIAnalyzer(config Config) (*OpenAIAnalyzer, error) {
	// Validate configuration
	if config.APIKey == "" || config.DeploymentName == "" || config.Endpoint == "" {
		return nil, fmt.Errorf("missing required OpenAI configuration")
	}

	// Create Azure OpenAI client
	keyCredential := azcore.NewKeyCredential(config.APIKey)
	client, err := azopenai.NewClientWithKeyCredential(config.Endpoint, keyCredential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	return &OpenAIAnalyzer{
		client: client,
		config: config,
	}, nil
}

// GenerateInsights generates AI-powered log analysis insights
func (oa *OpenAIAnalyzer) GenerateInsights(ctx context.Context, logAnalyzer *LogAnalyzer) (string, error) {
	// Prepare log texts with more context
	var criticalLogTexts []string
	var performanceLogTexts []string

	// Convert log entries to formatted strings
	for _, log := range logAnalyzer.criticalEvents {
		criticalLogTexts = append(criticalLogTexts,
			fmt.Sprintf("%s | %s | %s | %s",
				log.Timestamp, log.Namespace, log.PodName, log.LogContent,
			),
		)
	}

	for _, log := range logAnalyzer.performanceIssues {
		performanceLogTexts = append(performanceLogTexts,
			fmt.Sprintf("%s | %s | %s | %s",
				log.Timestamp, log.Namespace, log.PodName, log.LogContent,
			),
		)
	}

	// Include the existing detailed report for additional context
	detailedReport := logAnalyzer.generateDetailedReport()

	// Combine logs with additional context
	focusedLogs := fmt.Sprintf("Detailed Report:\n%s\n\nCritical Events:\n%s\n\nPerformance Issues:\n%s",
		detailedReport,
		strings.Join(criticalLogTexts, "\n"),
		strings.Join(performanceLogTexts, "\n"),
	)

	// Add a size check to prevent potential issues with very large inputs
	const maxInputSize = 10000 // Adjust based on your needs
	if len(focusedLogs) > maxInputSize {
		focusedLogs = focusedLogs[:maxInputSize]
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Prepare OpenAI request
	req := azopenai.ChatCompletionsOptions{
		Messages: []azopenai.ChatRequestMessageClassification{
			&azopenai.ChatRequestSystemMessage{
				Content: azopenai.NewChatRequestSystemMessageContent(AnalysisPrompt),
			},
			&azopenai.ChatRequestUserMessage{
				Content: azopenai.NewChatRequestUserMessageContent(
					fmt.Sprintf("Analyze the following Kubernetes log analysis and provide strategic insights and recommendations:\n\n%s", focusedLogs),
				),
			},
		},
		DeploymentName: &oa.config.DeploymentName,
		MaxTokens:      toInt32Ptr(750), // Increased token limit to prevent truncation
	}

	resp, err := oa.client.GetChatCompletions(ctx, req, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get chat completions: %w", err)
	}

	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
		return *resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no insights generated")
}

// Helper function to convert int to int32 pointer
func toInt32Ptr(i int) *int32 {
	int32Val := int32(i)
	return &int32Val
}
