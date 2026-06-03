package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Controller struct {
	k8sClient   *kubernetes.Clientset
	redisClient *redis.Client
	logger      *slog.Logger
}

func (c *Controller) Reconcile() {
	ctx := context.Background()
	// בדיקת גודל התור
	load, err := c.redisClient.LLen(ctx, "sunday_tasks_queue").Result()
	if err != nil {
		c.logger.Error("Failed to check Redis load", "error", err)
		return
	}

	// קבלת מצב ה-Worker
	scale, err := c.k8sClient.AppsV1().Deployments("default").GetScale(ctx, "sunday-worker", metav1.GetOptions{})
	if err != nil {
		c.logger.Error("Failed to get deployment scale", "error", err)
		return
	}

	// לוגיקת Scaling
	if load > 15 && scale.Spec.Replicas < 3 {
		scale.Spec.Replicas = 3
		_, err = c.k8sClient.AppsV1().Deployments("default").UpdateScale(ctx, "sunday-worker", scale, metav1.UpdateOptions{})
		if err == nil { c.logger.Info("Scaled UP", "load", load) }
	} else if load < 5 && scale.Spec.Replicas > 1 {
		scale.Spec.Replicas = 1
		_, err = c.k8sClient.AppsV1().Deployments("default").UpdateScale(ctx, "sunday-worker", scale, metav1.UpdateOptions{})
		if err == nil { c.logger.Info("Scaled DOWN", "load", load) }
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	config, _ := rest.InClusterConfig()
	k8sClient, _ := kubernetes.NewForConfig(config)

	c := &Controller{
		k8sClient:   k8sClient,
		redisClient: redis.NewClient(&redis.Options{Addr: "redis-service:6379"}),
		logger:      logger,
	}

	logger.Info("Operator ready (Polling Architecture)")

	// לולאת עבודה - סריקה כל 5 שניות
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			c.Reconcile()
		case <-quit:
			logger.Info("Shutting down gracefully")
			return
		}
	}
}