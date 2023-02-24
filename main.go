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

var (
	tlsCrt = "config/tls/tls.crt"
	tlsKey = "config/tls/tls.key"
)

func main() {

	mux := http.NewServeMux()

	// exec deploy pod name renew
	mux.HandleFunc("/mutation", deploy.AddAnno)

	server := &http.Server{
		Addr:        ":8443",
		Handler:     mux,
		ReadTimeout: 20 * time.Second, WriteTimeout: 20 * time.Second,
	}

	// add healthCheck
	go func() {
		healthCheck()
	}()

	log.Fatal(server.ListenAndServeTLS(tlsCrt, tlsKey))

}

func healthCheck() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health_check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}
