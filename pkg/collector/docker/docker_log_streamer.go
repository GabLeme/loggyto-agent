package docker

import (
	"bufio"
	"context"
	"fmt"
	"log"

	"log-agent/pkg/processor"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func StartLogStream(
	cli *client.Client,
	containerID string,
	logger *processor.LogProcessor,
	hostInfo map[string]string,
	cc *DockerCollector,
) {
	ctx := context.Background()
	logOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "10",
	}

	logReader, err := cli.ContainerLogs(ctx, containerID, logOptions)
	if err != nil {
		log.Printf("Error getting logs for docker container %s: %v", containerID, err)
		cc.mu.Lock()
		delete(cc.logTrackers, containerID)
		cc.mu.Unlock()
		return
	}
	defer logReader.Close()

	fmt.Printf("[Docker Container: %s] Streaming logs...\n", containerID[:12])

	scanner := bufio.NewScanner(logReader)
	for scanner.Scan() {
		logMessage := scanner.Bytes()

		if len(logMessage) > 8 {
			logMessage = logMessage[8:]
		}

		logger.ProcessLog("container", string(logMessage), map[string]string{
			"container_id": containerID,
			"host_name":    hostInfo["host_name"],
			"machine_ip":   hostInfo["machine_ip"],
			"os":           hostInfo["os"],
			"architecture": hostInfo["architecture"],
		})
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading logs for docker container %s: %v", containerID, err)
	}

	cc.mu.Lock()
	delete(cc.logTrackers, containerID)
	cc.mu.Unlock()
}
