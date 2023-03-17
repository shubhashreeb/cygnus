package main

import (
	"flag"
	"fmt"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"

	"github.com/sirupsen/logrus"
)

//GetClient returns a kubernetes client
func GetK8sClient() (*kubernetes.Clientset, error) {
	configpath := flag.String("kubeconfig", filepath.Join("/Users/shubhashree/", ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *configpath)
	if err != nil {
		logrus.Fatalf("Error occured while reading kubeconfig:%v", err)
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// Controller demonstrates how to implement a controller with client-go.
type Controller struct {
	// clinetset
	clientset *kubernetes.Clientset
	// informers for various resources
	podIndexer     cache.Indexer
	serviceIndexer cache.Indexer

	queue workqueue.RateLimitingInterface

	// informers for various resources
	podInformer     cache.Controller
	serviceInformer cache.Controller
}

// NewController creates a new Controller.
func NewController() *Controller {
	clientset, err := GetK8sClient()
	if err != nil {
		fmt.Println("Error in creating client")
		return nil
	}
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	return &Controller{
		clientset: clientset,
		queue:     queue,
	}
}

func (c *Controller) addPodInformer() *Controller {
	podListWatcher := cache.NewListWatchFromClient(c.clientset.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())

	// create the workqueue

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Pod than the version which was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(podListWatcher, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{

		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				fmt.Println("==========>>>>> POD Add event - %v", key)
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				fmt.Println("========>>>>> POD update event - %v", key)
				c.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				fmt.Println("=====>>>>>> POD delete event - %v", key)
				c.queue.Add(key)
			}
		},
	}, cache.Indexers{})
	c.podIndexer = indexer
	c.podInformer = informer
	return c
}

func (c *Controller) addServiceInformer() *Controller {
	fmt.Println("Adding service informer")
	podListWatcher := cache.NewListWatchFromClient(c.clientset.CoreV1().RESTClient(), "services", v1.NamespaceDefault, fields.Everything())
	indexer, informer := cache.NewIndexerInformer(podListWatcher, &v1.Service{}, 0, cache.ResourceEventHandlerFuncs{

		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				fmt.Println("==========>>>>> SERVICE Add event - %v", key)
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				fmt.Println("========>>>>> SERVICE update event - %v", key)
				c.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				fmt.Println("=====>>>>>> SERVICE delete event - %v", key)
				c.queue.Add(key)
			}
		},
	}, cache.Indexers{})
	c.serviceIndexer = indexer
	c.serviceInformer = informer
	return c
}

func (c *Controller) run() {
	stop := make(chan struct{})
	defer close(stop)
	go c.runPodInformer(stop)
	go c.runServiceInformer(stop)
	select {}
}

func (c *Controller) runPodInformer(stop chan struct{}) {
	c.podInformer.Run(stop)
}

func (c *Controller) runServiceInformer(stop chan struct{}) {
	c.serviceInformer.Run(stop)
}

func main() {
	controller := NewController()
	controller.addPodInformer()
	controller.addServiceInformer()
	controller.run()
}
