package internal

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ListPVCs returns all PVC objects in a given namespace
func ListPVCs(namespace string) ([]v1.PersistentVolumeClaim, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		return nil, err
	}
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pvcs.Items, nil
}
func GetPVC(namespace, pvcName string) (*corev1.PersistentVolumeClaim, error) {
	clientset, _ := GetK8sClient()
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
}

func ExecCommandInPod(clientset *kubernetes.Clientset, config *rest.Config, podName, namespace string, command []string) (string, error) {
	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: command,
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
			TTY:     false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", err
	}

	var stdout, stderr strings.Builder
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetUsedSizeInMBInPod executes du -sm inside a pod and returns used MB
func GetUsedSizeInMBInPod(clientset *kubernetes.Clientset, config *rest.Config, podName, namespace, mountPath string) (int64, error) {
	cmd := []string{"sh", "-c", fmt.Sprintf("du -sm %s 2>/dev/null || echo 0", mountPath)}

	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command: cmd,
			Stdout:  true,
			Stderr:  true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return 0, err
	}

	var stdout, stderr strings.Builder
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return 0, fmt.Errorf("exec failed: %v, stderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	fields := strings.Fields(output)
	if len(fields) == 0 {
		return 0, fmt.Errorf("unexpected du output: %q", output)
	}

	usedMB, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing du output: %w", err)
	}
	return usedMB, nil
}

// GetUsedSizeInMB sums used storage of PVC across all pods mounting it
// GetUsedSizeInMB sums used storage of a PVC across all pods mounting it
func GetUsedSizeInMB(clientset *kubernetes.Clientset, config *rest.Config, namespace, pvcName string) (int64, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	var totalUsed int64
	for _, pod := range pods.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == pvcName {
				// iterate containers
				for _, c := range pod.Spec.Containers {
					for _, m := range c.VolumeMounts {
						if m.Name == vol.Name {
							usedMB, err := execDuInPod(clientset, config, pod.Name, namespace, m.MountPath)
							if err == nil {
								totalUsed += usedMB
							}
						}
					}
				}
			}
		}
	}

	return totalUsed, nil
}

// execDuInPod executes `du -sm <mountPath>` in the pod and returns used MB
func execDuInPod(clientset *kubernetes.Clientset, config *rest.Config, podName, namespace, mountPath string) (int64, error) {
	cmd := []string{"sh", "-c", fmt.Sprintf("du -sm %s 2>/dev/null || echo 0", mountPath)}

	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		Param("container", "").
		Param("stdin", "false").
		Param("stdout", "true").
		Param("stderr", "true").
		Param("tty", "false").
		Param("command", cmd[0]).
		Param("command", cmd[1]).
		Param("command", cmd[2])

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return 0, err
	}

	var stdout, stderr strings.Builder
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return 0, fmt.Errorf("%v: %s", err, stderr.String())
	}

	outStr := strings.Fields(stdout.String())
	if len(outStr) == 0 {
		return 0, nil
	}

	usedMB, err := strconv.ParseInt(outStr[0], 10, 64)
	if err != nil {
		return 0, err
	}

	return usedMB, nil
}
