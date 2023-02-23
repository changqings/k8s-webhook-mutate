package main

import (
	"k8s-webhook-test/webhook"
	"log"
	"net/http"
	"time"
)

var (
	deploy = webhook.Deploy{
		Name:          "go-fiber",
		Namespace:     "default",
		PodNamePrefix: "eklet-",
	}
)

func main() {

	mux := http.NewServeMux()

	// add pod healthCheck
	mux.HandleFunc("/health_check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// exec deploy pod name renew
	mux.HandleFunc("/mutation", deploy.AddAnno)

	server := &http.Server{
		Addr:        ":8180",
		Handler:     mux,
		ReadTimeout: 20 * time.Second, WriteTimeout: 20 * time.Second,
	}

	log.Fatal(server.ListenAndServe())

}
