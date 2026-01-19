package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"  // הנה ה-Import שהיה חסר לך!
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var gvr = schema.GroupVersionResource{
	Group:    "sunday.com",
	Version:  "v1",
	Resource: "etherealpods",
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Ghost Operator is starting", "version", "v1.2", "env", "production")

	var config *rest.Config
	var err error

	// בדיקה האם רצים בתוך הקלאסטר או לוקאלית
	config, err = rest.InClusterConfig()
	if err != nil {
		slog.Info("Running outside of cluster, trying local kubeconfig")
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			kubeconfig = filepath.Join(os.Getenv("USERPROFILE"), ".kube", "config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			slog.Error("CRITICAL: Could not load Kubernetes config", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Info("Running inside Kubernetes cluster")
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		slog.Error("Failed to create dynamic client", "error", err)
		os.Exit(1)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error("Failed to create k8s client", "error", err)
		os.Exit(1)
	}

	slog.Info("Operator started successfully. Watching for EtherealPods...")

	for {
		list, err := dynamicClient.Resource(gvr).Namespace("default").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			slog.Error("Error listing custom resources", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, item := range list.Items {
			reconcile(item, k8sClient)
		}

		time.Sleep(5 * time.Second)
	}
}

// reconcile מקבלת כעת רק 2 פרמטרים
func reconcile(item unstructured.Unstructured, client *kubernetes.Clientset) {
	name := item.GetName()
	
	spec, found, _ := unstructured.NestedMap(item.Object, "spec")
	if !found {
		return
	}
	image, _, _ := unstructured.NestedString(spec, "image")
	if image == "" {
		image = "sunday-app:v1"
	}

	podName := "real-" + name

	_, err := client.CoreV1().Pods("default").Get(context.TODO(), podName, metav1.GetOptions{})

	if err != nil {
		slog.Info("Pod missing, resurrecting...", "pod", podName)
		createPod(client, podName, image)
	}
}

func createPod(client *kubernetes.Clientset, name string, image string) {
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"managed-by": "ethereal-operator", "app": "sunday-app"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "main-container",
					Image:           image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.FromInt(8080),
							},
						},
						InitialDelaySeconds: 10,
						PeriodSeconds:       15,
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	_, err := client.CoreV1().Pods("default").Create(context.TODO(), newPod, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to resurrect pod", "pod", name, "error", err)
	} else {
		slog.Info("Successfully resurrected pod", "pod", name)
	}
}