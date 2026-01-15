package main

import (
	"context"
	"log/slog" // לוגר מובנה בפורמט JSON (סטנדרט תעשייתי)
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	// הגדרת לוגר מקצועי - כל הלוגים יודפסו כ-JSON מוכן לניטור
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Ghost Operator is starting", "version", "v1.2", "env", "production")

	// הגדרת חיבור לקלאסטר
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		slog.Error("Failed to build kubeconfig", "error", err)
		os.Exit(1)
	}

	// יצירת קליינטים (Client-Go)
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

	// לופ ה-Reconciliation המרכזי
	for {
		list, err := dynamicClient.Resource(gvr).Namespace("default").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			slog.Error("Error listing custom resources", "error", err)
			time.Sleep(10 * time.Second)
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
	l := slog.With("resource", name) // הוספת הקשר ללוגים

	spec, found, _ := unstructured.NestedMap(ep.Object, "spec")
	if !found {
		l.Error("Spec not found in EtherealPod")
		return
	}

	imageName, _ := spec["image"].(string)
	podName := "real-" + name

	// בדיקת מצב הפוד הקיים
	existingPod, err := k8sClient.CoreV1().Pods("default").Get(context.TODO(), podName, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			l.Warn("Pod is missing! Reconciling state...", "action", "create")
			createPod(k8sClient, podName, imageName)
			updateStatus(dynClient, ep)
		} else {
			l.Error("Unexpected error fetching pod", "error", err)
		}
		return
	}

	// --- בדיקת Idempotency: האם הפוד תקין ותואם להגדרות? ---
	if existingPod.Spec.Containers[0].Image != imageName {
		l.Info("Pod configuration mismatch detected", "current_image", existingPod.Spec.Containers[0].Image, "desired_image", imageName)
		
		// מחיקה לצורך עדכון (בסיבוב הבא הוא ייוצר מחדש)
		err := k8sClient.CoreV1().Pods("default").Delete(context.TODO(), podName, metav1.DeleteOptions{})
		if err != nil {
			l.Error("Failed to delete outdated pod", "error", err)
		}
	}
}

func createPod(client *kubernetes.Clientset, name string, image string) {
	hostPathType := corev1.HostPathDirectoryOrCreate

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
					ImagePullPolicy: corev1.PullNever,
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
					VolumeMounts: []corev1.VolumeMount{
						{Name: "data-storage", MountPath: "/data"},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "data-storage",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/tmp/sunday-data",
							Type: &hostPathType,
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	_, err := client.CoreV1().Pods("default").Create(context.TODO(), newPod, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to create pod", "pod", name, "error", err)
	} else {
		slog.Info("Successfully created managed pod", "pod", name)
	}
}

func updateStatus(client *dynamic.DynamicClient, ep unstructured.Unstructured) {
	res, found, _ := unstructured.NestedInt64(ep.Object, "status", "resurrections")
	
	var nextRes int64 = 0
	if found {
		nextRes = res + 1
	}

	unstructured.SetNestedField(ep.Object, nextRes, "status", "resurrections")

	_, err := client.Resource(gvr).Namespace("default").UpdateStatus(context.TODO(), &ep, metav1.UpdateOptions{})
	if err != nil {
		slog.Error("Failed to sync resource status", "resource", ep.GetName(), "error", err)
	} else {
		slog.Info("Resource status synchronized", "resource", ep.GetName(), "resurrections", nextRes)
	}
}