package docker

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"log-agent/internal/config"
	"log-agent/internal/processor"
	"log-agent/internal/utils"

	"github.com/docker/docker/api/types/container"
	dockerevents "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

type DockerCollector struct {
	client      *client.Client
	stopChan    chan struct{}
	logTrackers map[string]struct{}
	mu          sync.Mutex
	Logger      *processor.LogProcessor
	hostInfo    map[string]string
	cfg         config.Config
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewContainerCollector(logger *processor.LogProcessor, cfg config.Config) *DockerCollector {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("[ERROR] Error creating Docker client: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &DockerCollector{
		client:      cli,
		stopChan:    make(chan struct{}),
		logTrackers: make(map[string]struct{}),
		Logger:      logger,
		hostInfo:    utils.GetHostMetadata(),
		cfg:         cfg,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (cc *DockerCollector) Start() {
	log.Println("[INFO] Docker Collector started...")
	go cc.watchDockerEvents()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cc.stopChan:
			log.Println("[WARNING] Stopping Docker Collector...")
			return
		case <-ticker.C:
			cc.collectContainers()
		}
	}
}

func (cc *DockerCollector) collectContainers() {
	ctx, cancel := context.WithTimeout(cc.ctx, 10*time.Second)
	defer cancel()

	containers, err := cc.client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		log.Printf("[ERROR] Error listing docker containers: %v", err)
		return
	}

	for _, cont := range containers {
		containerID := cont.ID
		containerName := strings.TrimPrefix(cont.Names[0], "/")

		if cc.shouldIgnore(containerName) {
			continue
		}

		cc.mu.Lock()
		if _, exists := cc.logTrackers[containerID]; exists {
			cc.mu.Unlock()
			continue
		}
		cc.logTrackers[containerID] = struct{}{}
		cc.mu.Unlock()

		go cc.startContainerLogStream(containerID, containerName)
	}
}

func (cc *DockerCollector) shouldIgnore(containerName string) bool {
	for _, ignored := range cc.cfg.IgnoredContainers {
		if ignored == containerName {
			log.Printf("[INFO] Ignoring container %s", containerName)
			return true
		}
	}
	return false
}

func (cc *DockerCollector) startContainerLogStream(containerID, containerName string) {
	ctx, cancel := context.WithTimeout(cc.ctx, 10*time.Second)
	defer cancel()

	StartLogStream(ctx, cc.client, containerID, containerName, cc.Logger, cc)
}

func (cc *DockerCollector) Stop() {
	log.Println("[INFO] Stopping Docker Collector...")
	cc.cancel()
	close(cc.stopChan)
	cc.client.Close()
}

func (cc *DockerCollector) watchDockerEvents() {
	for {
		select {
		case <-cc.stopChan:
			log.Println("[INFO] Stopping Docker event watcher...")
			return
		default:
			cc.streamDockerEvents()
			time.Sleep(5 * time.Second)
		}
	}
}

func (cc *DockerCollector) streamDockerEvents() {
	ctx, cancel := context.WithCancel(cc.ctx)
	defer cancel()

	eventChan, errChan := cc.client.Events(ctx, dockerevents.ListOptions{})

	for {
		select {
		case <-cc.stopChan:
			return
		case event := <-eventChan:
			cc.handleDockerEvent(event)
		case err := <-errChan:
			log.Printf("[ERROR] Error watching Docker events: %v", err)
			return
		}
	}
}

func (cc *DockerCollector) handleDockerEvent(event dockerevents.Message) {
	if event.Type != "container" {
		return
	}

	containerID := event.Actor.ID
	containerName := event.Actor.Attributes["name"]

	switch event.Action {
	case "start":
		cc.handleContainerStarted(containerID, containerName)
	case "die", "stop", "kill":
		cc.handleContainerStopped(containerID, containerName, string(event.Action))
	}
}

func (cc *DockerCollector) handleContainerStarted(containerID, containerName string) {
	if cc.shouldIgnore(containerName) {
		return
	}

	cc.mu.Lock()
	if _, exists := cc.logTrackers[containerID]; exists {
		cc.mu.Unlock()
		return
	}
	cc.logTrackers[containerID] = struct{}{}
	cc.mu.Unlock()

	log.Printf("[INFO] Detected new container: %s. Starting log stream...", containerName)
	go cc.startContainerLogStream(containerID, containerName)
}

func (cc *DockerCollector) handleContainerStopped(containerID, containerName, reason string) {
	log.Printf("[WARNING] Container %s (%s) has stopped. Reason: %s", containerName, containerID[:12], reason)
	cc.mu.Lock()
	delete(cc.logTrackers, containerID)
	cc.mu.Unlock()
}
