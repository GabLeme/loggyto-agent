package inputs

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubernetesCollector struct {
	clientset   *kubernetes.Clientset
	stopChan    chan struct{}
	logTrackers sync.Map
}

func NewKubernetesCollector() *KubernetesCollector {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating in-cluster Kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	return &KubernetesCollector{
		clientset: clientset,
		stopChan:  make(chan struct{}),
	}
}

func (kc *KubernetesCollector) Start() {
	fmt.Println("Kubernetes Collector started...")

	for {
		select {
		case <-kc.stopChan:
			fmt.Println("Stopping Kubernetes Collector...")
			return
		default:
			pods, err := kc.clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Printf("Error listing pods: %v", err)
				continue
			}

			for _, pod := range pods.Items {
				if _, loaded := kc.logTrackers.LoadOrStore(pod.Name, true); !loaded {
					go kc.streamLogs(pod.Namespace, pod.Name)
				}
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func (kc *KubernetesCollector) streamLogs(namespace, podName string) {
	ctx := context.Background()
	logOptions := &v1.PodLogOptions{
		Follow:    true,
		TailLines: func(i int64) *int64 { return &i }(10),
	}

	logRequest := kc.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	logStream, err := logRequest.Stream(ctx)
	if err != nil {
		log.Printf("Error getting logs for pod %s/%s: %v", namespace, podName, err)
		kc.logTrackers.Delete(podName)
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
			log.Printf("Error reading logs for pod %s/%s: %v", namespace, podName, err)
			return
		}
		fmt.Printf("[Pod: %s] %s", podName, string(buf[:bytesRead]))
	}

	kc.logTrackers.Delete(podName)
}

func (kc *KubernetesCollector) Stop() {
	close(kc.stopChan)
}
