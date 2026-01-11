package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr" // Added for Liveness Probe
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	gvr = schema.GroupVersionResource{
		Group:    "sunday.com",
		Version:  "v1",
		Resource: "etherealpods",
	}
)

func main() {
	fmt.Println("👻 Ethereal Operator is starting...")

	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	for {
		list, err := dynamicClient.Resource(gvr).Namespace("default").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("Error listing: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, item := range list.Items {
			reconcile(item, k8sClient, dynamicClient)
		}
		time.Sleep(5 * time.Second)
	}
}

func reconcile(ep unstructured.Unstructured, k8sClient *kubernetes.Clientset, dynClient *dynamic.DynamicClient) {
	name := ep.GetName()
	spec, found, _ := unstructured.NestedMap(ep.Object, "spec")
	if !found {
		return
	}
	imageName, _ := spec["image"].(string)

	podName := "real-" + name
	_, err := k8sClient.CoreV1().Pods("default").Get(context.TODO(), podName, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Printf("⚠️  Pod %s is missing! Resurrecting...\n", podName)

			// 1. Resurrect the pod
			createPod(k8sClient, podName, imageName)

			// 2. Update status in CRD
			updateStatus(dynClient, ep)

		} else {
			fmt.Printf("Error getting pod: %v\n", err)
		}
	}
}

func createPod(client *kubernetes.Clientset, name string, image string) {
	hostPathType := corev1.HostPathDirectoryOrCreate

	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"app": "sunday-app"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            "main-container",
					Image:           image,
					ImagePullPolicy: corev1.PullNever,
					// Liveness Probe to detect deadlocks
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.FromInt(8080),
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       10,
					},
					// Mount data directory
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "sunday-storage",
							MountPath: "/data",
						},
					},
				},
			},
			// Configure HostPath for persistence
			Volumes: []corev1.Volume{
				{
					Name: "sunday-storage",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/tmp/sunday-storage",
							Type: &hostPathType,
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
	client.CoreV1().Pods("default").Create(context.TODO(), newPod, metav1.CreateOptions{})
	fmt.Printf("✨ Pod resurrected with PERSISTENCE: %s\n", name)
}

// updateStatus increments the resurrection counter in the Custom Resource
// updateStatus handles the logic: 
// If it's the first run -> Set to 0.
// If it's a crash -> Increment by 1.
func updateStatus(client *dynamic.DynamicClient, ep unstructured.Unstructured) {
    // מנסים לקרוא את הסטטוס הקיים
    currentRestarts, found, _ := unstructured.NestedInt64(ep.Object, "status", "resurrections")
    
    var newRestarts int64

    if !found {
        // אם הסטטוס לא קיים בכלל - סימן שזו הפעם הראשונה שהפוד נוצר
        // אז אנחנו רק מאתחלים ל-0 (ולא מוסיפים 1)
        newRestarts = 0
        fmt.Println("🌱 Initializing Pod for the first time (Restarts = 0)")
    } else {
     
        newRestarts = currentRestarts + 1
        fmt.Printf("🔄 Crash detected! Incrementing restarts to %d\n", newRestarts)
    }


    unstructured.SetNestedField(ep.Object, newRestarts, "status", "resurrections")

    _, err := client.Resource(gvr).Namespace("default").UpdateStatus(context.TODO(), &ep, metav1.UpdateOptions{})
    if err != nil {
        fmt.Printf("Failed to update status: %v\n", err)
    }
}