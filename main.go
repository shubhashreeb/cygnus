package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"

	"github.com/sirupsen/logrus"
)

// GetClient returns a kubernetes client
func GetK8sClient() (*kubernetes.Clientset, error) {
	configpath := flag.String("kubeconfig", filepath.Join("/Users/shubhashree/", ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	if configpath == nil {
		return nil, fmt.Errorf("Invalid file path")
	}
	config, err := clientcmd.BuildConfigFromFlags("", *configpath)
	if err != nil {
		logrus.Fatalf("Error occured while reading kubeconfig:%v", err)
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// Controller demonstrates how to implement a controller with client-go.
type Controller struct {
	// clientset
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

func (c *Controller) sync() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// extract the key and event type from the queue

	keyValue := key.(EventInfo)

	t := reflect.TypeOf(key)
	fmt.Println("Key type is ", t)

	switch keyValue.eventObj {
	case POD:
		// Invoke the method containing the business logic
		err := c.syncPodToStdout(keyValue.key, keyValue.eventType)
		// Handle the error if something went wrong during the execution of the business logic
		c.handleErr(err, key)
		return true
	case SERVICE:
		// Invoke the method containing the business logic
		err := c.syncSvcToStdout(keyValue.key,
			keyValue.eventType)
		// Handle the error if something went wrong during the execution of the business logic
		c.handleErr(err, key)
		return true
	}
	return false
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the pod to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *Controller) syncPodToStdout(key string, eventType eventType) error {
	obj, exists, err := c.podIndexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		fmt.Printf("Pod %s does not exist anymore\n", key)
		return nil
	}

	// Note that you also have to check the uid if you have a local controlled resource, which
	// is dependent on the actual instance, to detect that a Pod was recreated with the same name
	switch eventType {
	case Add:
		fmt.Println("Add")
		fmt.Printf("Add for Pod %s\n", obj.(*v1.Pod).GetName())
		if len(obj.(*v1.Pod).Status.ContainerStatuses) < 1 {
			fmt.Println("Pod container status is un available")
		} else {
			fmt.Printf("Pod waiting status - %v, Running status - %v",
				obj.(*v1.Pod).Status.ContainerStatuses[0].State.Waiting,
				obj.(*v1.Pod).Status.ContainerStatuses[0].State.Running)
		}
	case Update:
		fmt.Println("Update")
		fmt.Printf("Sync/Update for Pod %s\n", obj.(*v1.Pod).GetName())
		if len(obj.(*v1.Pod).Status.ContainerStatuses) < 1 {
			fmt.Println("Pod container status is un available")
		} else {
			fmt.Printf("Pod waiting status - %v, Running status - %v",
				obj.(*v1.Pod).Status.ContainerStatuses[0].State.Waiting,
				obj.(*v1.Pod).Status.ContainerStatuses[0].State.Running)
		}
	case Delete:
		fmt.Println("Delete")
		fmt.Printf("Delete for Pod %s\n", obj.(*v1.Pod).GetName())
	}
	return nil
}

func (c *Controller) syncSvcToStdout(key string, eventType eventType) error {
	obj, exists, err := c.serviceIndexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Pod, so that we will see a delete for one pod
		fmt.Printf("Pod %s does not exist anymore\n", key)
		return nil
	}
	switch eventType {
	case Add:
		fmt.Println("Processing Add service")
		fmt.Printf("Add for SVC %s\n", obj.(*v1.Service).GetName())
		if v, ok := obj.(*v1.Service).GetAnnotations()["cygnus.io/path"]; ok {
			fmt.Printf("***** Add for SVC annotations path is %s\n", v)
		}
		fmt.Printf("Add for SVC annotations %s\n", obj.(*v1.Service).GetAnnotations())

	case Update:
		fmt.Println(" Processing Update service")
		fmt.Printf("Update for SVC %s\n", obj.(*v1.Service).GetName())
		if v, ok := obj.(*v1.Service).GetAnnotations()["cygnus.io/path"]; ok {
			fmt.Printf("***** Update for SVC annotations path is %s\n", v)
		}
		fmt.Printf("Update for SVC annotations %s\n", obj.(*v1.Service).GetAnnotations())

	case Delete:
		fmt.Println(" Processing Delete service ")
		fmt.Printf("Delete for SVC %s\n", obj.(*v1.Service).GetName())
	}
	// Note that you also have to check the uid if you have a local controlled resource, which
	// is dependent on the actual instance, to detect that a Pod was recreated with the same name

	// if len(obj.(*v1.Pod).Status.ContainerStatuses) < 1 {
	// 	fmt.Println("Pod container status is un available")
	// } else {
	// 	fmt.Printf("Pod waiting status - %v, Running status - %v",
	// 		obj.(*v1.Pod).Status.ContainerStatuses[0].State.Waiting,
	// 		obj.(*v1.Pod).Status.ContainerStatuses[0].State.Running)
	// }
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	klog.Infof("Dropping pod %q out of the queue: %v", key, err)
}

type eventObj int

const (
	POD eventObj = iota
	SERVICE
)

type eventType int

const (
	Add eventType = iota
	Update
	Delete
)

type EventInfo struct {
	eventObj  eventObj
	key       string
	eventType eventType
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
				e := EventInfo{
					eventObj:  POD,
					key:       key,
					eventType: Add,
				}
				fmt.Println("==========>>>>> POD Add event - %v", key)
				c.queue.Add(e)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				e := EventInfo{
					eventObj:  POD,
					key:       key,
					eventType: Update,
				}
				fmt.Println("========>>>>> POD update event - %v", key)
				c.queue.Add(e)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				e := EventInfo{
					eventObj:  POD,
					key:       key,
					eventType: Delete,
				}
				fmt.Println("=====>>>>>> POD delete event - %v", key)
				c.queue.Add(e)
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
				e := EventInfo{
					eventObj:  SERVICE,
					key:       key,
					eventType: Add,
				}
				fmt.Println("==========>>>>> SERVICE Add event - %v", key)
				c.queue.Add(e)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				e := EventInfo{
					eventObj:  SERVICE,
					key:       key,
					eventType: Update,
				}
				fmt.Println("========>>>>> SERVICE update event - %v", key)
				c.queue.Add(e)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				e := EventInfo{
					eventObj:  SERVICE,
					key:       key,
					eventType: Delete,
				}
				fmt.Println("=====>>>>>> SERVICE delete event - %v", key)
				c.queue.Add(e)
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

	go c.sync()
	select {}
}

func (c *Controller) runPodInformer(stop chan struct{}) {
	c.podInformer.Run(stop)
}

func (c *Controller) runServiceInformer(stop chan struct{}) {
	c.serviceInformer.Run(stop)
}

func main() {
	controller := NewController().
		addPodInformer().
		addServiceInformer()
	controller.run()
}
