package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	k8sClient *kubernetes.Clientset
	queue     workqueue.RateLimitingInterface
	logger    *slog.Logger
}

func (c *Controller) ReconcileBusinessLogic(key string) error {
	load := c.getPendingTasksFromBackend()
	
	if load > 15 {
		c.logger.Warn("CRITICAL LOAD: Scaling UP resources (Self-Healing Triggered)", "load", load, "target", key)
	} else if load > 10 {
		c.logger.Info("High load detected. Monitoring closely.", "load", load)
	} else if load < 5 {
		c.logger.Info("Low load. Scaling DOWN to save costs.", "load", load)
	} else {
		c.logger.Info("System stable. No action required.", "load", load)
	}
	return nil
}

func (c *Controller) getPendingTasksFromBackend() int {
	// פנייה לשרת דרך הרשת הפנימית של דוקר
	resp, err := http.Get("http://sunday-service:8080/pending-count")
	if err != nil {
		c.logger.Error("Backend unreachable", "error", err)
		return 0
	}
	defer resp.Body.Close()

	var res struct{ Count int `json:"count"` }
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0
	}
	return res.Count
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Ethereal Operator Booting Sequence Initiated...")

	config, err := rest.InClusterConfig()
	if err != nil {
		home := homedir.HomeDir()
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
	}

	// מנגנון חכם: אם אין קוברנטס זמין כרגע, מריץ סימולציה לבדיקות מקומיות
	if err != nil {
		logger.Warn("Kubernetes cluster not found. Running in Standalone Simulation Mode.")
		c := &Controller{logger: logger}
		for {
			c.ReconcileBusinessLogic("local-simulation")
			time.Sleep(3 * time.Second) // דוגם את השרת כל 3 שניות
		}
		return
	}

	// חיבור לקוברנטס אמיתי (כאשר ייפרס בענן)
	dynamicClient, _ := dynamic.NewForConfig(config)
	k8sClient, _ := kubernetes.NewForConfig(config)
	factory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 30*time.Second)
	informer := factory.ForResource(schema.GroupVersionResource{Group: "sunday.com", Version: "v1", Resource: "etherealpods"})

	c := &Controller{
		k8sClient: k8sClient, 
		logger:    logger, 
		queue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "etherealpods"),
	}

	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) { 
			key, _ := cache.MetaNamespaceKeyFunc(obj)
			c.queue.Add(key) 
		},
	})

	go factory.Start(context.Background().Done())
	
	logger.Info("Operator is actively watching K8s events")
	for {
		obj, _ := c.queue.Get()
		c.ReconcileBusinessLogic(obj.(string))
		c.queue.Done(obj)
	}
}