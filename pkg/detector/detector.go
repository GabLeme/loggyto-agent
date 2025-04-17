package detector

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"log-agent/pkg/collector/docker"
	"log-agent/pkg/collector/journald"
	"log-agent/pkg/collector/kubernetes"
	"log-agent/pkg/outputs"
	"log-agent/pkg/processor"
)

func DetectJournald() bool {
	paths := []string{
		"/run/systemd/journal/socket",
		"/var/run/systemd/journal/socket",
		"/run/log/journal",
		"/var/log/journal",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			fmt.Println("[INFO] Detected systemd journald at:", path)
			return true
		}
	}
	fmt.Println("[WARNING] journald not detected on this system.")
	return false
}

func DetectEnvironment() string {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		fmt.Println("Detected: Kubernetes Cluster")
		return "kubernetes"
	}

	if _, err := exec.LookPath("microk8s"); err == nil {
		cmd := exec.Command("microk8s", "status")
		output, _ := cmd.Output()
		if string(output) != "" {
			fmt.Println("Detected: MicroK8s")
			return "microk8s"
		}
	}

	if _, err := exec.LookPath("minikube"); err == nil {
		cmd := exec.Command("minikube", "status")
		output, _ := cmd.Output()
		if string(output) != "" {
			fmt.Println("Detected: Minikube")
			return "minikube"
		}
	}

	if DetectJournald() {
		return "journald"
	}

	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		fmt.Println("Detected: Docker")
		return "docker"
	}

	fmt.Println("Environment detection failed: Running in an unknown environment")
	return "unknown"
}

func StartCollector() {
	env := DetectEnvironment()

	output := outputs.NewStdoutOutput()
	logProcessor := processor.NewLogProcessor(output)

	var collector interface {
		Start()
		Stop()
	}

	switch env {
	case "kubernetes", "microk8s", "minikube":
		k8sCollector := kubernetes.NewKubernetesCollector(logProcessor)
		collector = k8sCollector
	case "docker":
		collector = docker.NewContainerCollector(logProcessor)
	case "journald":
		collector = journald.NewJournaldCollector(logProcessor)
	default:
		collector = journald.NewJournaldCollector(logProcessor)
		return
	}

	fmt.Println("Starting collector...")
	go collector.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down collector...")
	collector.Stop()
}
