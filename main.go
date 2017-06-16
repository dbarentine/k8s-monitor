package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"reflect"
	"encoding/json"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/pkg/api"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"strings"
	"k8s.io/client-go/rest"
)

type LogMessage struct {
	ObjectType string
	ObjectName string
	EventType string
	AssignedNode string
	Replicas int32
	Timestamp time.Time
}

func main() {
	clientset := createClientSet()

	createWatcher(clientset.CoreV1().RESTClient(), &v1.Pod{}, "pods")
	createWatcher(clientset.CoreV1().RESTClient(), &v1.Service{}, "services")
	createWatcher(clientset.CoreV1().RESTClient(), &v1.Secret{}, "secrets")
	createWatcher(clientset.CoreV1().RESTClient(), &v1.ConfigMap{}, "configmaps")
	createWatcher(clientset.CoreV1().RESTClient(), &v1.Namespace{}, "namespaces")
	createWatcher(clientset.CoreV1().RESTClient(), &v1.ReplicationController{}, "replicationcontrollers")
	createWatcher(clientset.ExtensionsV1beta1().RESTClient(), &v1beta1.ReplicaSet{}, "replicasets")
	createWatcher(clientset.ExtensionsV1beta1().RESTClient(), &v1beta1.Ingress{}, "ingresses")
	createWatcher(clientset.ExtensionsV1beta1().RESTClient(), &v1beta1.Deployment{}, "deployments")

	for{
		time.Sleep(time.Second)
	}
}

func createWatcher(c cache.Getter, obj runtime.Object, resource string) cache.Controller {
	watchlist := cache.NewListWatchFromClient(c, resource, api.NamespaceAll, fields.Everything())
	resyncPeriod := 30 * time.Minute

	_, controller := cache.NewInformer(
		watchlist,
		obj,
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				toJson(obj, "Created")
			},
			DeleteFunc: func(obj interface{}) {
				toJson(obj, "Deleted")
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				toJson(newObj, "Updated")
			},
		},
	)

	go controller.Run(wait.NeverStop)
	return controller
}

func toJson(obj interface{}, eventType string) string {
	objType := reflect.TypeOf(obj)

	logMessage := &LogMessage{ObjectType: strings.TrimLeft(objType.String(), "*"), ObjectName: getName(obj), EventType: eventType, Timestamp: time.Now()}

	//Add additional information to log message
	switch t := obj.(type) {
	case *v1.Pod:
		logMessage.AssignedNode = t.Spec.NodeName
	case *v1beta1.Deployment:
		logMessage.Replicas = *t.Spec.Replicas
	case *v1beta1.ReplicaSet:
		logMessage.Replicas = *t.Spec.Replicas
	}

	b, err := json.Marshal(logMessage)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	jsonString := string(b)
	fmt.Println(jsonString)
	return string(jsonString)
}

func getName(obj interface{}) string {
	switch t := obj.(type) {
	default:
		return "Unknown Name"
	case *v1.Namespace:
		return t.ObjectMeta.Name
	case *v1.Pod:
		return t.ObjectMeta.Name
	case *v1beta1.Deployment:
		return t.ObjectMeta.Name
	case *v1beta1.DaemonSet:
		return t.ObjectMeta.Name
	case *v1beta1.ReplicaSet:
		return t.ObjectMeta.Name
	case *v1.Secret:
		return t.ObjectMeta.Name
	case *v1.ConfigMap:
		return t.ObjectMeta.Name
	case *v1beta1.Ingress:
		return t.ObjectMeta.Name
	case *v1.Service:
		return t.ObjectMeta.Name
	}

	return "Unknown Name"
}

func createClientSet() *kubernetes.Clientset {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	//Default to using kubeconfig or commandline arg
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		// kubeconfig failed attempt in cluster config
		// creates the in-cluster config
		fmt.Println("Creating in cluster configuration...")
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	return buildClientSet(config)
}

func buildClientSet(config *rest.Config) *kubernetes.Clientset {
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}