package internal

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FindPodsForPVC returns a list of pod names that mount the given PVC in a namespace
func FindPodsForPVC(namespace, pvcName string) ([]string, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		return nil, err
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, pod := range pods.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == pvcName {
				result = append(result, pod.Name)
			}
		}
	}
	return result, nil
}
func ListPods(namespace string) ([]corev1.Pod, error) {
	client, _ := GetK8sClient()
	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing pods in %s: %v", namespace, err)
	}
	return podList.Items, nil
}

// FindPodAndMountPathForPVC returns the first pod and mount path that is using the PVC
func FindPodAndMountPathForPVC(namespace, pvcName string) (string, string, error) {
	clientset, err := GetK8sClient()
	if err != nil {
		return "", "", err
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", "", err
	}

	for _, pod := range pods.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == pvcName {
				// find the mount path from the first container
				for _, container := range pod.Spec.Containers {
					for _, vm := range container.VolumeMounts {
						if vm.Name == vol.Name {
							return pod.Name, vm.MountPath, nil
						}
					}
				}
			}
		}
	}
	return "", "", fmt.Errorf("no pod found using PVC %s", pvcName)
}

// ExecInPod executes a command in a pod container and returns stdout as string
func ExecInPod(podName, namespace, command string) (string, error) {
	cmd := exec.Command("kubectl", "exec", "-n", namespace, podName, "--", "sh", "-c", command)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, stderr.String())
	}
	return out.String(), nil
}
