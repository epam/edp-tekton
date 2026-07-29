package collector

import (
	"bytes"
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// PodLogFetcher fetches container logs through the Kubernetes pods/log
// subresource. Tail limiting is applied server-side via PodLogOptions.
type PodLogFetcher struct {
	client kubernetes.Interface
}

func NewPodLogFetcher(client kubernetes.Interface) *PodLogFetcher {
	return &PodLogFetcher{client: client}
}

func (f *PodLogFetcher) GetLogs(
	ctx context.Context,
	namespace, podName, container string,
	tailLines int64,
) (string, error) {
	opts := &corev1.PodLogOptions{Container: container}
	if tailLines > 0 {
		opts.TailLines = &tailLines
	}

	stream, err := f.client.CoreV1().Pods(namespace).GetLogs(podName, opts).Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to stream logs of pod %s container %s: %w", podName, container, err)
	}

	defer func() {
		_ = stream.Close()
	}()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, stream); err != nil {
		return "", fmt.Errorf("failed to read logs of pod %s container %s: %w", podName, container, err)
	}

	return buf.String(), nil
}
