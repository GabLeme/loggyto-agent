package docker

import (
	"context"
	"log"
	"sync"
	"time"

	"log-agent/pkg/processor"
	"log-agent/pkg/utils"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerCollector struct {
	client      *client.Client
	stopChan    chan struct{}
	logTrackers map[string]struct{}
	mu          sync.Mutex
	Logger      *processor.LogProcessor
	hostInfo    map[string]string
}

func NewContainerCollector(logger *processor.LogProcessor) *DockerCollector {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	return &DockerCollector{
		client:      cli,
		stopChan:    make(chan struct{}),
		logTrackers: make(map[string]struct{}),
		Logger:      logger,
		hostInfo:    utils.GetHostMetadata(),
	}
}

func (cc *DockerCollector) Start() {
	log.Println("Docker Collector started...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cc.stopChan:
			log.Println("Stopping Docker Collector...")
			return
		case <-ticker.C:
			cc.collectContainers()
		}
	}
}

func (cc *DockerCollector) collectContainers() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	containers, err := cc.client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		log.Printf("Error listing docker containers: %v", err)
		return
	}

	for _, container := range containers {
		cc.mu.Lock()
		if _, exists := cc.logTrackers[container.ID]; !exists {
			cc.logTrackers[container.ID] = struct{}{}
			go StartLogStream(cc.client, container.ID, cc.Logger, cc.hostInfo, cc)
		}
		cc.mu.Unlock()
	}
}

func (cc *DockerCollector) Stop() {
	close(cc.stopChan)
	cc.client.Close()
}
