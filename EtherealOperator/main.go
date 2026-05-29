package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	ctx := context.TODO()
	deploymentName := "sunday-app"
	// לוגיקה עסקית אקטיבית
	if load > 15 {
		c.logger.Warn("CRITICAL LOAD: Scaling UP resources via K8s API", "load", load)
		c.updateDeploymentScale(ctx, deploymentName, 3)
	} else if load < 5 {
		c.logger.Info("Low load. Scaling DOWN via K8s API to save costs.", "load", load)
		c.updateDeploymentScale(ctx, deploymentName, 1)
	} else {
		c.logger.Info("System stable. No action required.", "load", load)
	}
	return nil
}

// updateDeploymentScale פונה ישירות ל-API של קוברנטס ומשנה את כמות השרתים בפועל
func (c *Controller) updateDeploymentScale(ctx context.Context, deploymentName string, desiredReplicas int32) {
	// הגנה: אם מורץ מקומית ללא קלאסטר (מצב סימולציה)
	if c.k8sClient == nil {
		c.logger.Info("Simulation Mode Active: K8s client is nil. Would dynamically scale to", "replicas", desiredReplicas)
		return
	}

	scale, err := c.k8sClient.AppsV1().Deployments("default").GetScale(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		c.logger.Error("Failed to fetch current deployment scale from API", "error", err)
		return
	}

	// שינוי מצב (Mutation) רק אם יש צורך
	if scale.Spec.Replicas != desiredReplicas {
		scale.Spec.Replicas = desiredReplicas
		_, err = c.k8sClient.AppsV1().Deployments("default").UpdateScale(ctx, deploymentName, scale, metav1.UpdateOptions{})
		if err != nil {
			c.logger.Error("API UpdateScale request failed", "error", err)
		} else {
			c.logger.Info("Successfully mutated K8s state: Deployment scale updated", "new_replicas", desiredReplicas)
		}
	}
}

func (c *Controller) getPendingTasksFromBackend() int {
	resp, err := http.Get("http://sunday-service:8080/pending-count")
	if err != nil {
		c.logger.Error("Network Error: Failed to reach backend", "error", err) // לוג שגיאת תקשורת
		return 0
	}
	defer resp.Body.Close()

	var res struct {
		Count int `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		c.logger.Error("JSON Error: Failed to parse response", "error", err) // לוג שגיאת נתונים
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

	// הפעלת מצב סימולציה אם המערכת רצה מחוץ לקלאסטר
	if err != nil {
		logger.Warn("Kubernetes cluster not found. Running in Standalone Simulation Mode.")
		c := &Controller{logger: logger}
		for {
			c.ReconcileBusinessLogic("local-simulation")
			time.Sleep(3 * time.Second)
		}
		return
	}

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
		DeleteFunc: func(obj interface{}) {
			logger.Warn("Resource deletion detected via Informer! Processing logic...")
			key, _ := cache.MetaNamespaceKeyFunc(obj)
			c.queue.Add(key)
		},
	})

	go factory.Start(context.Background().Done())

	logger.Info("Operator is actively processing cluster events")
	for {
		obj, _ := c.queue.Get()
		c.ReconcileBusinessLogic(obj.(string))
		c.queue.Done(obj)
	}
}
