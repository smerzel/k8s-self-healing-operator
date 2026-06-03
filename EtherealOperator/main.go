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
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/workqueue"
)

type Controller struct {
	k8sClient   *kubernetes.Clientset
	redisClient *redis.Client
	queue       workqueue.RateLimitingInterface
	logger      *slog.Logger
}

func (c *Controller) Reconcile() {
	ctx := context.Background()
	load, err := c.redisClient.LLen(ctx, "sunday_tasks_queue").Result()
	if err != nil {
		c.logger.Error("Failed to check Redis load", "error", err)
		return
	}

	scale, err := c.k8sClient.AppsV1().Deployments("default").GetScale(ctx, "sunday-worker", metav1.GetOptions{})
	if err != nil {
		c.logger.Error("Failed to get deployment scale", "error", err)
		return
	}

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
		queue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "etherealpods"),
	}

	// הגדרת ה-Informer לשינויים ב-Deployments
	factory := informers.NewSharedInformerFactory(k8sClient, time.Minute*10)
	deploymentInformer := factory.Apps().V1().Deployments().Informer()

	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.queue.Add("reconcile-event")
		},
	})

	stopCh := make(chan struct{})
	factory.Start(stopCh)
	
	// האזנה לאירועי Redis
	pubsub := c.redisClient.Subscribe(context.Background(), "task_events")
	go func() {
		for range pubsub.Channel() {
			c.queue.Add("task-event")
		}
	}()

	logger.Info("Operator ready (Full Event-Driven Architecture)")

	// לולאת עיבוד אירועים חכמה
	go func() {
		for {
			obj, _ := c.queue.Get()
			c.Reconcile()
			c.queue.Done(obj)
		}
	}()

	// ניהול סגירה תקינה (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	close(stopCh)
	logger.Info("Shutting down gracefully")
}