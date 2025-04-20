package detector

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"log-agent/internal/collector/docker"
	"log-agent/internal/sender"

	// "log-agent/internal/collector/journald"
	"log-agent/internal/collector/kubernetes"
	"log-agent/internal/outputs"
	"log-agent/internal/processor"
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
			log.Printf("[INFO] Detected systemd journald at: %s", path)
			return true
		}
	}
	log.Printf("[INFO] Journald not detected on this system.")
	return false
}

func DetectEnvironment() string {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		log.Println("[INFO] Detected Kubernetes environment.")
		return "kubernetes"
	}

	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		log.Println("[INFO] Detected Docker environment.")
		return "docker"
	}

	if DetectJournald() {
		return "journald"
	}

	log.Println("[WARNING] Environment detection failed. Running in unknown environment.")
	return "unknown"
}

func StartCollector() {
	env := DetectEnvironment()
	cfg := sender.LoadConfigFromEnv()
	s := sender.NewSender(cfg)

	output := outputs.NewStdoutOutput()
	logProcessor := processor.NewLogProcessor(s, output)

	var collector interface {
		Start()
		Stop()
	}

	switch env {
	case "kubernetes":
		collector = kubernetes.NewKubernetesCollector(logProcessor)
	case "docker":
		collector = docker.NewContainerCollector(logProcessor)
	case "journald":
		// collector = journald.NewJournaldCollector(logProcessor)
	default:
		log.Println("[ERROR] No compatible environment detected. Exiting.")
		return
	}

	log.Println("[INFO] Starting collector...")
	go collector.Start()

	// graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("[INFO] Shutting down collector...")
	collector.Stop()
}
