package kube

import (
	"context"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func CreatePod(clientset *kubernetes.Clientset, podName string, nameSpace string) error {
	podSepc := &v1.Pod{
		TypeMeta: v12.TypeMeta{},
		ObjectMeta: v12.ObjectMeta{
			Name:      podName,
			Namespace: nameSpace,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: v1.PodSpec{Containers: []v1.Container{
			{
				Name:    "test-container",
				Image:   "ubuntu",
				Command: []string{"sleep", "1d"},
			},
		},
		},
		Status: v1.PodStatus{},
	}
	pod, err := clientset.CoreV1().Pods(nameSpace).Create(context.Background(), podSepc, v12.CreateOptions{})
	if err != nil {
		klog.Error("Error creating pod ", err)
		return err
	}
	klog.Info("Pod created successfully ", pod.Name)
	return nil
}

func DeleteAllTestPods(clientset *kubernetes.Clientset, nameSpace string) error {
	err := clientset.CoreV1().Pods(nameSpace).DeleteCollection(context.Background(), v12.DeleteOptions{}, v12.ListOptions{LabelSelector: "app"})
	if err != nil {
		klog.Error("Error deleting pods ", err)
	}
	return err
}

func DeletePod(clientset *kubernetes.Clientset, podName string, nameSpace string) error {
	err := clientset.CoreV1().Pods(nameSpace).Delete(context.Background(), podName, v12.DeleteOptions{})
	if err != nil {
		klog.Error("Error deleting pods ", err)
	}
	return err
}
