package main

import (
	"encoding/json"
	"log"
	"net/http"

	"context"

	"github.com/gorilla/mux"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type event struct {
	ID          string `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

type allEvents []event

var events = allEvents{
	{
		ID:          "1",
		Title:       "Introduction to Golang",
		Description: "Come join us for a chance to learn how golang works and get to eventually try it out",
	},
}

func getOneEvent(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["id"]

	for _, singleEvent := range events {
		if singleEvent.ID == eventID {
			json.NewEncoder(w).Encode(singleEvent)
		}
	}
}

func getAllEvents(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(events)
}

type ContainerImage struct {
	Name  string `json:"Name"`
	Image string `json:"Image"`
}

type object struct {
	Name       string `json:"Name"`
	Type       string `json:"Type"`
	Containers []ContainerImage
}

type allObjects []object

func k8s(w http.ResponseWriter, r *http.Request) {

	// Out cluster config
	// kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	// log.Println("Using kubeconfig file: ", kubeconfig)
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// In cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create an rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	deployments, err := clientset.AppsV1().Deployments("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalln("failed to get deployments:", err)
	}

	var allPods allObjects

	for _, deployment := range deployments.Items {
		var containers []ContainerImage
		for _, container := range deployment.Spec.Template.Spec.Containers {
			var spec = ContainerImage{
				Name:  container.Name,
				Image: container.Image,
			}
			containers = append(containers, spec)
		}
		var podobj = object{
			Name:       deployment.GetName(),
			Type:       "Deployment",
			Containers: containers,
		}
		allPods = append(allPods, podobj)
		// json.NewEncoder(w).Encode(deployment)
		// fmt.Fprintf(w, "IIIIIIICCCCIII %s\n", deployment.Spec.Template.Spec.Containers[0].Image)
	}
	json.NewEncoder(w).Encode(allPods)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", k8s)
	router.HandleFunc("/events", getAllEvents).Methods("GET")
	router.HandleFunc("/events/{id}", getOneEvent).Methods("GET")
	router.HandleFunc("/k8s", k8s).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
