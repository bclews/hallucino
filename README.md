# Intelligent Kubernetes Log Analyser

The **Intelligent Kubernetes Log Analyser** (nicknamed `Hallucino`) is a command-line tool designed to analyse Kubernetes logs, leveraging LLMs to extract insights, summarise patterns, and identify anomalies.

I created this project during our two-day "engineering development days" event, intended to build skills, explore interests, and connect with our colleagues. This was my chance to flex some advanced Go muscles, dive deep into concurrency, and master memory management while working on something practical and fun.

## The Wins

I am happy with how this project turned out! I definitely leaned on LLMs to scaffold this project, and get me to the learning bits faster. So, a few highlights:

- Kubernetes Wizardry: Built a seamless integration with Kubernetes to fetch and process logs like a pro.
- LLM Integration: Used our on-premise OpenAI Azure service to add AI-powered insights to log analysis.
- Concurrency Mastery: Optimised log parsing to handle large volumes efficiently, without sacrificing performance or memory.
- CLI Design: Crafted a sleek command-line tool with Cobra that’s both powerful and user-friendly.
- Learning and Growing: Honed my Go skills, explored Kubernetes in depth, and gained a bit more of understanding of AI applications.

## Lessons Learned

Not everything was smooth sailing, of course. Debugging concurrency issues? Yikes. Concurrency is hard, but it’s also incredibly powerful. I learned a lot about Goroutines, channels, and sync.WaitGroup.

Tweaking LLms to handle unstructured log data? Let’s just say it was a learning experience. But that’s the point, right?

## Features

- **Concurrent Log Parsing**: Efficiently retrieves and processes logs from multiple Kubernetes pods and containers using Goroutines.
- **Log Analysis with AI**: summarises logs, detects common error patterns, and provides actionable insights using Azure's OpenAI.
- **Customizable Output**: Supports raw log printing or Markdown-rendered summaries.
- **Robust Kubernetes Integration**: Seamlessly interacts with Kubernetes clusters to fetch logs and container details.
- **Structured Logging**: Built with the `zap` library for performance and readability.

## Project Structure

```
.
├── cmd
│   └── root.go            # Command-line interface definition
├── go.mod                 # Module dependencies
├── go.sum                 # Dependency checksums
├── hallucino              # Binary output directory
├── internal
│   ├── analysis           # Analysis engine for logs
│   │   ├── analyser.go    # Core log analysis logic
│   │   └── openai.go      # Integration with Llama or Azure OpenAI
│   ├── k8s                # Kubernetes API interactions
│   │   └── client.go      # Pod and container log retrieval
│   ├── logger             # Custom logger configuration
│   │   └── logger.go      # `zap`-based logger setup
│   └── storage            # Log storage and management
│       └── storage.go     # Thread-safe log handling
└── main.go                # Entry point for the application
```

## Installation

1. Clone this repository:

   ```bash
   git clone https://github.com/<your-username>/intelligent-log-analyser.git
   cd intelligent-log-analyser
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Build the application:

   ```bash
   go build -o hallucino
   ```

## Usage

```bash
Usage:
  hallucino [flags]

Flags:
      --container string    Specific container name
  -h, --help                help for hallucino
      --kubeconfig string   Path to kubeconfig file
      --namespace string    Kubernetes namespace
      --pod string          Specific pod name
      --print-raw           Pretty print retrieved logs

```

### CLI Flags

- `--kubeconfig` : Path to the Kubernetes configuration file (optional).
- `--namespace`  : Kubernetes namespace to query (default: `default`).
- `--pod`        : Pod name for log retrieval (optional).
- `--container`  : Container name within the pod (optional).
- `--printRaw`   : Print raw logs instead of AI-processed summaries (optional).

## How It Works

1. **Kubernetes Log Retrieval**:  
   The tool fetches logs using the Kubernetes client-go library, supporting specific pods and containers or all containers within a namespace.

2. **Concurrent Processing**:  
   Logs are processed in parallel to enhance performance and minimise memory bottlenecks.

3. **AI-Powered Insights**:  
   Logs are analysed using an LLM (e.g., Azure OpenAI) to summarise patterns, identify anomalies, and provide actionable recommendations.

4. **Reporting**:  
   Insights are rendered as Markdown and printed to the terminal using the Glamour library for enhanced readability.

## Example Output

### Raw Logs

```
2024-11-27T10:00:00Z [pod1-container1] ERROR: Connection timeout.
2024-11-27T10:01:00Z [pod1-container2] WARN: High memory usage detected.
```

### summarised Insights

```markdown
## Kubernetes Log Analysis

### Errors
- Connection timeout in `pod1-container1`.

### Warnings
- High memory usage detected in `pod1-container2`.

### Recommendations
- Review connection stability for `pod1-container1`.
- Investigate memory-intensive processes in `pod1-container2`.
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
