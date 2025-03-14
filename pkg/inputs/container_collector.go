package inputs

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"log-agent/pkg/processor"
	"log-agent/pkg/utils"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ContainerCollector struct {
	client      *client.Client
	stopChan    chan struct{}
	logTrackers sync.Map
	Logger      *processor.LogProcessor
	hostInfo    map[string]string
}

func NewContainerCollector(logger *processor.LogProcessor) *ContainerCollector {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	return &ContainerCollector{
		client:   cli,
		stopChan: make(chan struct{}),
		Logger:   logger,
		hostInfo: utils.GetHostMetadata(),
	}
}

func (cc *ContainerCollector) Start() {
	fmt.Println("Container Collector started...")

	for {
		select {
		case <-cc.stopChan:
			fmt.Println("Stopping Container Collector...")
			return
		default:
			containers, err := cc.client.ContainerList(context.Background(), container.ListOptions{})
			if err != nil {
				log.Printf("Error listing containers: %v", err)
				continue
			}

			for _, container := range containers {
				if _, loaded := cc.logTrackers.LoadOrStore(container.ID, true); !loaded {
					go cc.streamLogs(container.ID)
				}
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func (cc *ContainerCollector) streamLogs(containerID string) {
	ctx := context.Background()
	logOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "10",
	}

	logReader, err := cc.client.ContainerLogs(ctx, containerID, logOptions)
	if err != nil {
		log.Printf("Error getting logs for container %s: %v", containerID, err)
		cc.logTrackers.Delete(containerID)
		return
	}
	defer logReader.Close()

	fmt.Printf("[Container: %s] Streaming logs...\n", containerID[:12])

	scanner := bufio.NewScanner(logReader)
	for scanner.Scan() {
		logMessage := scanner.Bytes()

		if len(logMessage) > 8 {
			logMessage = logMessage[8:]
		}

		cc.Logger.ProcessLog("container", string(logMessage), map[string]string{
			"container_id": containerID,
			"host_name":    cc.hostInfo["host_name"],
			"machine_ip":   cc.hostInfo["machine_ip"],
			"os":           cc.hostInfo["os"],
			"architecture": cc.hostInfo["architecture"],
		})
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading logs for container %s: %v", containerID, err)
	}

	cc.logTrackers.Delete(containerID)
}

func (cc *ContainerCollector) Stop() {
	close(cc.stopChan)
}
