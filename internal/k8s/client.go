package k8s

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type LogEntry struct {
	Namespace  string
	PodName    string
	Container  string
	LogContent string
	Timestamp  string
}

// ListPods retrieves all pod names in a given namespace
func ListPods(client *kubernetes.Clientset, namespace string) ([]string, error) {
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var podNames []string
	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames, nil
}

// ListContainers retrieves all container names for a specific pod
func ListContainers(client *kubernetes.Clientset, namespace, podName string) ([]string, error) {
	pod, err := client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var containerNames []string
	for _, container := range pod.Spec.Containers {
		containerNames = append(containerNames, container.Name)
	}

	return containerNames, nil
}

// RetrievePodLogs retrieves logs for a specific pod and container
func RetrievePodLogs(client *kubernetes.Clientset, namespace, podName, containerName string) ([]LogEntry, error) {
	req := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
	})

	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error opening log stream: %v", err)
	}
	defer podLogs.Close()

	// Read logs
	var logs []LogEntry
	logBytes, err := io.ReadAll(podLogs)
	if err != nil {
		return nil, fmt.Errorf("error reading logs: %v", err)
	}

	// Parse logs into entries
	logLines := strings.Split(string(logBytes), "\n")
	for _, line := range logLines {
		if line == "" {
			continue
		}
		logs = append(logs, LogEntry{
			Namespace:  namespace,
			PodName:    podName,
			Container:  containerName,
			LogContent: line,
			Timestamp:  time.Now().Format(time.RFC3339),
		})
	}

	return logs, nil
}
