package inputs

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
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
	nodeName    string
	namespace   string
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

	collector := &KubernetesCollector{
		clientset: clientset,
		stopChan:  make(chan struct{}),
		namespace: getNamespace(),
	}

	collector.nodeName = collector.getCurrentNodeName()
	fmt.Printf("Agent running on node: %s\n", collector.nodeName)

	return collector
}

func getNamespace() string {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		log.Fatalf("Error reading namespace file: %v", err)
	}
	return string(data)
}

func (kc *KubernetesCollector) getCurrentNodeName() string {
	podName := kc.getPodName()

	fieldSelector := fmt.Sprintf("metadata.name=%s", podName)
	podList, err := kc.clientset.CoreV1().Pods(kc.namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil || len(podList.Items) == 0 {
		log.Fatalf("Error getting current pod: %v", err)
	}

	return podList.Items[0].Spec.NodeName
}

func (kc *KubernetesCollector) getPodName() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Error getting hostname: %v", err)
	}
	return hostname
}

func (kc *KubernetesCollector) Start() {
	fmt.Println("Kubernetes Collector started...")

	excludedNamespaces := map[string]bool{
		"kube-system":   true,
		"istio-system":  true,
		"monitoring":    true,
		"calico-system": true,
		"logging":       true,
		"cilium-system": true,
	}

	ownPodName := kc.getPodName()

	for {
		select {
		case <-kc.stopChan:
			fmt.Println("Stopping Kubernetes Collector...")
			return
		default:
			fieldSelector := fmt.Sprintf("spec.nodeName=%s", kc.nodeName)
			pods, err := kc.clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
				FieldSelector: fieldSelector,
			})
			if err != nil {
				log.Printf("Error listing pods: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, pod := range pods.Items {
				if excludedNamespaces[pod.Namespace] {
					continue
				}

				if pod.Name == ownPodName {
					continue
				}

				podKey := fmt.Sprintf("%s-%s", pod.Namespace, pod.UID)

				if _, loaded := kc.logTrackers.LoadOrStore(podKey, true); !loaded {
					go kc.streamLogs(pod.Namespace, pod.Name, podKey)
				}
			}

			time.Sleep(5 * time.Second)
		}
	}
}

func (kc *KubernetesCollector) streamLogs(namespace, podName, podKey string) {
	ctx := context.Background()
	logOptions := &v1.PodLogOptions{
		Follow:    true,
		TailLines: func(i int64) *int64 { return &i }(10),
	}

	logRequest := kc.clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	logStream, err := logRequest.Stream(ctx)
	if err != nil {
		log.Printf("Error getting logs for pod %s/%s: %v", namespace, podName, err)
		kc.logTrackers.Delete(podKey)
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
			break
		}
		fmt.Printf("[Pod: %s] %s", podName, string(buf[:bytesRead]))
	}

	kc.logTrackers.Delete(podKey)
}

func (kc *KubernetesCollector) Stop() {
	close(kc.stopChan)
}
