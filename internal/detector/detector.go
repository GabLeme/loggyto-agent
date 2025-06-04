package detector

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"log-agent/internal/collector/docker"
	"log-agent/internal/collector/journald"
	"log-agent/internal/collector/kubernetes"
	"log-agent/internal/config"
	"log-agent/internal/outputs"
	"log-agent/internal/processor"
	"log-agent/internal/sender"
)

type Collector interface {
	Start()
	Stop()
}

func DetectDocker() bool {
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		log.Println("[INFO] Detected Docker environment.")
		return true
	}
	return false
}

func DetectKubernetes() bool {
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount"); err == nil {
		log.Println("[INFO] Detected Kubernetes environment.")
		return true
	}
	return false
}

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
	log.Println("[INFO] Journald not detected on this system.")
	return false
}

func StartCollectors() {
	cfg := config.LoadConfigFromEnv()
	s := sender.NewSender(cfg)
	output := outputs.NewStdoutOutput()
	logProcessor := processor.NewLogProcessor(s, output)

	var collectors []Collector

	if DetectDocker() {
		collectors = append(collectors, docker.NewContainerCollector(logProcessor, cfg))
	}

	if DetectKubernetes() {
		collectors = append(collectors, kubernetes.NewKubernetesCollector(logProcessor))
	}

	if DetectJournald() {
		collectors = append(collectors, journald.NewJournaldCollector(logProcessor))
	}

	if len(collectors) == 0 {
		log.Println("[ERROR] No compatible environments detected. Exiting.")
		return
	}

	log.Printf("[INFO] Starting %d collectors...", len(collectors))
	for _, c := range collectors {
		go c.Start()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("[INFO] Shutting down collectors...")
	for _, c := range collectors {
		c.Stop()
	}
}
