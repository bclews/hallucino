package cmd

import (
	"context"
	"fmt"
	"hallucino/internal/analysis"
	"hallucino/internal/k8s"
	"hallucino/internal/storage"
	"os"
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig string
	namespace  string
	pod        string
	container  string
	printRaw   bool
	logger     *zap.Logger
	logStore   *storage.LogStorage
)

var rootCmd = &cobra.Command{
	Use:           "hallucino",
	Short:         "Kubernetes Log Retrieval Tool",
	Long:          "A CLI tool to retrieve logs from Kubernetes clusters with advanced filtering and storage capabilities",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger
		logger, err := zap.NewProduction()
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		defer logger.Sync()

		// Validate input combinations
		if err := validateInputCombinations(namespace, pod, container); err != nil {
			return err
		}

		// Initialize log storage
		logStore = storage.NewLogStorage()

		// Create Kubernetes client
		client, err := createK8sClient()
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}

		// Retrieve logs based on input
		if err := retrieveLogs(client); err != nil {
			return fmt.Errorf("log retrieval failed: %w", err)
		}

		// Pretty print logs if print-raw flag is set
		if printRaw {
			logStore.PrettyPrintLogs()
		} else {
			// Analyze logs
			if err := analyzeKubernetsLogs(logStore); err != nil {
				return fmt.Errorf("log analysis failed: %w", err)
			}
		}

		return nil
	},
}

func validateInputCombinations(namespace, pod, container string) error {
	// If no parameters are specified, return an error with usage instructions
	if namespace == "" && pod == "" && container == "" {
		return fmt.Errorf(
			`no parameters specified. Please provide at least a namespace.

Usage examples:
  hallucino --kubeconfig=/path/to/config --namespace my-namespace
  hallucino --kubeconfig=/path/to/config --namespace my-namespace --pod my-pod
  hallucino --kubeconfig=/path/to/config --namespace my-namespace --pod my-pod --container my-container`,
		)
	}

	// Case 1: Container specified without pod or namespace
	if container != "" && (pod == "" || namespace == "") {
		return fmt.Errorf(
			"container must be specified with both a pod and a namespace. For example:\n" +
				"  --namespace my-namespace --pod my-pod --container my-container",
		)
	}

	// Case 2: Pod specified without namespace
	if pod != "" && namespace == "" {
		return fmt.Errorf(
			"pod must be specified with a namespace. For example:\n" +
				"  --namespace my-namespace --pod my-pod",
		)
	}

	return nil
}

func createK8sClient() (*kubernetes.Clientset, error) {
	// Use provided kubeconfig or default
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}

	// Load Kubernetes configuration
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes config: %v", err)
	}

	// Create Kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client: %v", err)
	}

	return client, nil
}

func retrieveLogs(client *kubernetes.Clientset) error {
	// Retrieve logs based on specified parameters
	var pods []string
	var wg sync.WaitGroup
	logChan := make(chan k8s.LogEntry, 100)
	errorChan := make(chan error, 10)

	// Determine pods to retrieve logs from
	if pod == "" {
		// If no specific pod, get all pods in namespace
		podList, err := k8s.ListPods(client, namespace)
		if err != nil {
			return fmt.Errorf("failed to list pods: %v", err)
		}
		pods = podList
	} else {
		pods = []string{pod}
	}

	// Concurrent log retrieval
	for _, podName := range pods {
		wg.Add(1)
		go func(podName string) {
			defer wg.Done()

			// Determine containers
			var containers []string
			if container != "" {
				containers = []string{container}
			} else {
				// Get all containers in the pod
				podContainers, err := k8s.ListContainers(client, namespace, podName)
				if err != nil {
					errorChan <- fmt.Errorf("failed to list containers for pod %s: %v", podName, err)
					return
				}
				containers = podContainers
			}

			// Retrieve logs for each container
			for _, containerName := range containers {
				wg.Add(1)
				go func(podName, containerName string) {
					defer wg.Done()
					logs, err := k8s.RetrievePodLogs(client, namespace, podName, containerName)
					if err != nil {
						errorChan <- fmt.Errorf("failed to retrieve logs for pod %s, container %s: %v",
							podName, containerName, err)
						return
					}

					// Send logs to channel
					for _, log := range logs {
						logChan <- log
					}
				}(podName, containerName)
			}
		}(podName)
	}

	// Close channels when done
	go func() {
		wg.Wait()
		close(logChan)
		close(errorChan)
	}()

	// Process logs and errors with pretty printing
	var totalLogs int
	var logsProcessed sync.WaitGroup
	logsProcessed.Add(1)

	go func() {
		defer logsProcessed.Done()
		for {
			select {
			case log, ok := <-logChan:
				if !ok {
					// Logs channel closed
					return
				}

				// Store log
				logStore.AddLog(log)
				totalLogs++
			case err, ok := <-errorChan:
				if !ok {
					// Error channel closed
					break
				}
				// Print errors in red
				color.Red("Error: %v", err)
			}
		}
	}()

	// Wait for log processing to complete
	logsProcessed.Wait()

	return nil
}

func analyzeKubernetsLogs(logStorage *storage.LogStorage) error {
	// Get logs from storage
	logs := logStorage.GetLogs()

	// Create log analyzer
	logAnalyzer := analysis.NewLogAnalyzer(logs)

	// Create OpenAI analyzer
	openaiConfig := analysis.Config{
		APIKey:         os.Getenv("AZURE_API_KEY"),
		Endpoint:       os.Getenv("AZURE_API_BASE"),
		DeploymentName: os.Getenv("AZURE_DEPLOYMENT_NAME"),
	}

	openaiAnalyzer, err := analysis.NewOpenAIAnalyzer(openaiConfig)
	if err != nil {
		return fmt.Errorf("failed to create OpenAI analyzer: %w", err)
	}

	// Generate insights
	insights, err := openaiAnalyzer.GenerateInsights(context.Background(), logAnalyzer)
	if err != nil {
		return fmt.Errorf("failed to generate insights: %w", err)
	}

	// Print or process insights
	out, err := glamour.Render(insights, "dark")
	if err != nil {
		fmt.Println("Error rendering markdown:", err)
	} else {
		fmt.Println(out)
	}

	return nil
}

func init() {
	rootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	rootCmd.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace")
	rootCmd.Flags().StringVar(&pod, "pod", "", "Specific pod name")
	rootCmd.Flags().StringVar(&container, "container", "", "Specific container name")
	rootCmd.Flags().BoolVar(&printRaw, "print-raw", false, "Pretty print retrieved logs")
}

// Execute adds all child commands to the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		fmt.Fprintln(os.Stderr, rootCmd.UsageString()) // Optionally show usage on error
		os.Exit(1)
	}
}
