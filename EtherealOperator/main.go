package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/redis/go-redis/v9"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/workqueue"
)

type Controller struct {
	k8sClient *kubernetes.Clientset
	queue     workqueue.RateLimitingInterface
	logger    *slog.Logger
}

func (c *Controller) ReconcileBusinessLogic(key string) error {
	load, err := c.getPendingTasksFromBackend()
	if err != nil {
		c.logger.Error("Skipping reconciliation due to backend communication failure", "error", err)
		return err
	}

	ctx := context.TODO()
	deploymentName := "sunday-worker"
	
	// שליפת המצב הנוכחי כדי לא להציף את ה-API בבקשות מיותרות
	scale, err := c.k8sClient.AppsV1().Deployments("default").GetScale(ctx, deploymentName, metav1.GetOptions{})
	currentReplicas := int32(1)
	if err == nil {
		currentReplicas = scale.Spec.Replicas
	}

	if load > 15 {
		if currentReplicas < 3 {
			c.logger.Warn("CRITICAL LOAD: Scaling UP WORKERS", "load", load)
			c.updateDeploymentScale(ctx, deploymentName, 3)
		}
	} else if load < 5 {
		if currentReplicas > 1 {
			c.logger.Info("Low load. Scaling DOWN WORKERS to save costs.", "load", load)
			c.updateDeploymentScale(ctx, deploymentName, 1)
		}
	} else {
		c.logger.Info("System stable.", "load", load)
	}
	return nil
}

func (c *Controller) updateDeploymentScale(ctx context.Context, deploymentName string, desiredReplicas int32) {
	scale, err := c.k8sClient.AppsV1().Deployments("default").GetScale(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		c.logger.Error("Failed to fetch scale", "error", err)
		return
	}
	scale.Spec.Replicas = desiredReplicas
	_, err = c.k8sClient.AppsV1().Deployments("default").UpdateScale(ctx, deploymentName, scale, metav1.UpdateOptions{})
	if err != nil {
		c.logger.Error("UpdateScale failed", "error", err)
	} else {
		c.logger.Info("Deployment scale updated", "new_replicas", desiredReplicas)
	}
}

func (c *Controller) getPendingTasksFromBackend() (int, error) {
	resp, err := http.Get("http://sunday-service:8080/pending-count")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var res struct{ Count int `json:"count"` }
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Count, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	config, _ := rest.InClusterConfig()
	k8sClient, _ := kubernetes.NewForConfig(config)

	c := &Controller{
		k8sClient: k8sClient,
		logger:    logger,
		queue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "etherealpods"),
	}

	// חיבור ל-Redis להאזנה לאירועים (Event-Driven)
	redisClient := redis.NewClient(&redis.Options{Addr: "redis-service:6379"})
	pubsub := redisClient.Subscribe(context.Background(), "task_events")

	go func() {
		for range pubsub.Channel() {
			c.queue.Add("task-pushed-event")
		}
	}()

	logger.Info("Operator ready (Push Architecture)")
	for {
		obj, _ := c.queue.Get()
		if err := c.ReconcileBusinessLogic("task-event"); err != nil {
			c.queue.AddRateLimited(obj)
		} else {
			c.queue.Forget(obj)
		}
		c.queue.Done(obj)
	}
}