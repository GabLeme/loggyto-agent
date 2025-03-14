package inputs

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type ContainerCollector struct {
	client      *client.Client
	stopChan    chan struct{}
	logTrackers sync.Map
}

func NewContainerCollector() *ContainerCollector {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	return &ContainerCollector{
		client:   cli,
		stopChan: make(chan struct{}),
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
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, logReader)
	if err != nil {
		log.Printf("Error processing logs for container %s: %v", containerID, err)
	}

	cc.logTrackers.Delete(containerID)
}

func (cc *ContainerCollector) Stop() {
	close(cc.stopChan)
}
