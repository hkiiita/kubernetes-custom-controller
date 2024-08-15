package main

import (
	"flag"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"kubernetesController/controller"
	"kubernetesController/kube"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	namespace := flag.String("namespace", "default", "kubernetes namespace where pods will be created.")

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	go func() {
		count := 1
		for true {
			time.Sleep(1 * time.Second)
			_ = kube.CreatePod(clientset, strconv.Itoa(count), *namespace)
			count = count + 1
			if count >= 10 {
				for i := 1; i <= count; i++ {
					time.Sleep(10 * time.Second)
					_ = kube.DeletePod(clientset, strconv.Itoa(i), *namespace)
				}
			}
		}
	}()

	//------------------------ controller -----------------
	ch := make(chan struct{})
	defer close(ch)
	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 5*time.Second, informers.WithNamespace("default"))
	podInformer := factory.Core().V1().Pods().Informer()

	factory.Start(ch)

	c := controller.NewController(clientset, podInformer, *namespace)
	c.Run(ch)

}
