package controller

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	v12 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"kubernetesController/kube"
	"strings"
	"time"
)

type CustomController struct {
	clientSet      kubernetes.Interface
	queue          workqueue.RateLimitingInterface
	podLister      v12.PodLister
	podCacheSynced cache.InformerSynced
	nameSpace      string
}

func NewController(clientset kubernetes.Interface, podInformer cache.SharedIndexInformer, nameSpace string) *CustomController {
	c := &CustomController{
		clientSet:      clientset,
		queue:          workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		podLister:      nil,
		podCacheSynced: podInformer.HasSynced,
		nameSpace:      nameSpace,
	}

	_, err := podInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			UpdateFunc: nil,
			DeleteFunc: c.handleDelete,
		})
	if err != nil {
		klog.Info("Error adding event funcs to handler ", err)
		return nil
	}

	return c

}

func (c *CustomController) Run(stopper chan struct{}) {
	klog.Info("***** Starting Custom Controller *******")
	if !cache.WaitForCacheSync(stopper, c.podCacheSynced) {
		klog.Error("Timeout waiting for caches to sync.")
		return
	} else {
		klog.Info("cache sync was successfull.")
	}
	wait.Until(c.worker, 2*time.Second, stopper)

	<-stopper
	klog.Info("---- Stopping controller ---- ")

}

func (c *CustomController) worker() {
	item, shutDown := c.queue.Get()
	if shutDown {
		klog.Info("-- Queue empty ---")
		return
	}
	defer c.queue.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		klog.Error("Error getting key from cache ", err)
	}

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		klog.Error("Error splitting key into name and namespace ", err)
		return
	}

	_, err = c.clientSet.CoreV1().Pods(c.nameSpace).Get(context.Background(), name, v13.GetOptions{})

	if errors.IsNotFound(err) {
		// Pod deleted hence delete its sibling pod
		if strings.Contains(name, "sibling") {
			return
		}
		err := kube.DeletePod(c.clientSet.(*kubernetes.Clientset), name+"-sibling", c.nameSpace)
		if err != nil {
			klog.Error("Error deleting sibling pod ", err)
			return
		}
		return
	}

	if err != nil {
		klog.Error("Error while checking presence of pod ", err)
		return
	}

	//create a sibling pod
	if strings.Contains(name, "sibling") {
		return
	}
	err = kube.CreatePod(c.clientSet.(*kubernetes.Clientset), name+"-sibling", c.nameSpace)
	if err != nil {
		return
	}

}

func (c *CustomController) handleAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	klog.Info("Pod creation event  : ", pod.Name)
	c.queue.Add(obj)
}

func (c *CustomController) handleDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	klog.Info("Pod deletion event  ", pod.Name)
	c.queue.Add(obj)
}
