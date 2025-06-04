package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"log-agent/internal/processor"
)

func StartLogStream(
	ctx context.Context,
	cli *client.Client,
	containerID string,
	containerName string,
	logger *processor.LogProcessor,
	cc *DockerCollector,
) {
	defer cc.removeLogTracker(containerID)

	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		log.Printf("[ERROR] Error inspecting container %s: %v", containerID, err)
		return
	}

	state := containerJSON.State
	if state == nil || !state.Running {
		log.Printf("[WARNING] Container %s is not running (status: %s), skipping log stream.", containerID, state.Status)
		return
	}

	if containerJSON.HostConfig.LogConfig.Type != "json-file" {
		log.Printf("[ERROR] Unsupported log driver for container %s: %s", containerID, containerJSON.HostConfig.LogConfig.Type)
		return
	}

	logOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "0",
	}

	var logReader io.ReadCloser
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		logReader, err = cli.ContainerLogs(ctx, containerID, logOptions)
		if err == nil {
			break
		}
		log.Printf("[WARNING] Retrying to get logs for container %s (%d/%d)...", containerID, i+1, maxRetries)
		time.Sleep(time.Duration(2<<i) * time.Millisecond)
	}
	if err != nil {
		log.Printf("[ERROR] Failed to get logs for container %s after retries: %v", containerID, err)
		return
	}
	defer logReader.Close()

	enrichedMeta := map[string]string{
		"container_id":   containerID,
		"container_name": containerName,
	}
	for k, v := range containerJSON.Config.Labels {
		enrichedMeta[fmt.Sprintf("label_%s", k)] = v
	}

	if containerJSON.Config.Tty {
		readStream(logReader, containerID, containerName, logger, enrichedMeta)
	} else {
		stdoutReader, stdoutWriter := io.Pipe()
		stderrReader, stderrWriter := io.Pipe()

		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			defer wg.Done()
			readStream(stdoutReader, containerID, containerName, logger, enrichedMeta)
		}()

		go func() {
			defer wg.Done()
			readStream(stderrReader, containerID, containerName, logger, enrichedMeta)
		}()

		go func() {
			defer wg.Done()
			_, err := stdcopy.StdCopy(stdoutWriter, stderrWriter, logReader)
			if err != nil {
				log.Printf("[ERROR] Error reading multiplexed logs for %s: %v", containerID, err)
			}
			stdoutWriter.Close()
			stderrWriter.Close()
		}()

		wg.Wait()
	}

	if group, ok := logger.Flush(containerID); ok && group != nil {
		logger.ProcessLog(containerName, group.Message, enrichedMeta)
	}
}

func readStream(reader io.Reader, containerID, containerName string, logger *processor.LogProcessor, metadata map[string]string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		logMessage := scanner.Text()
		logger.ProcessLog(containerName, logMessage, metadata)
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("Error reading logs for container %s: %v", containerID, err)
	}
}

func (cc *DockerCollector) removeLogTracker(containerID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	delete(cc.logTrackers, containerID)
}
