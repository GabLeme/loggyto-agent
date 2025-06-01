package kubernetes

import (
	"context"
	"io"
	"log"

	"log-agent/internal/processor"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type KubernetesLogStreamer struct {
	clientset *kubernetes.Clientset
	logger    *processor.LogProcessor
	namespace string
	podName   string
	podKey    string
	hostInfo  map[string]string
}

func NewKubernetesLogStreamer(
	clientset *kubernetes.Clientset,
	logger *processor.LogProcessor,
	namespace,
	podName,
	podKey string,
	hostInfo map[string]string,
) *KubernetesLogStreamer {
	return &KubernetesLogStreamer{
		clientset: clientset,
		logger:    logger,
		namespace: namespace,
		podName:   podName,
		podKey:    podKey,
		hostInfo:  hostInfo,
	}
}

func (kls *KubernetesLogStreamer) StreamLogs() {
	ctx := context.Background()
	logOptions := &v1.PodLogOptions{
		Follow:    true,
		TailLines: func(i int64) *int64 { return &i }(10),
	}

	logRequest := kls.clientset.CoreV1().Pods(kls.namespace).GetLogs(kls.podName, logOptions)
	logStream, err := logRequest.Stream(ctx)
	if err != nil {
		log.Printf("Error getting logs for pod %s/%s: %v", kls.namespace, kls.podName, err)
		return
	}
	defer logStream.Close()

	buf := make([]byte, 4096)
	for {
		bytesRead, err := logStream.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading logs for pod %s/%s: %v", kls.namespace, kls.podName, err)
			break
		}

		logMessage := string(buf[:bytesRead])

		kls.logger.ProcessLog("kubernetes", logMessage, map[string]string{
			"pod_name":  kls.podName,
			"namespace": kls.namespace,
		})
	}
}
