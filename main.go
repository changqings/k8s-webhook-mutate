package main

import (
	"encoding/base64"
	"k8s-webhook-test/webhook"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	if keyDecodeFile, err := base64FileToStringFile(tlsKey); err == nil {
		tlsKey = keyDecodeFile
	}
	if crtDecodeFile, err := base64FileToStringFile(tlsCrt); err == nil {
		tlsCrt = crtDecodeFile
	}

	log.Println("Mute http server start running on port :8443 ...")
	log.Fatal(server.ListenAndServeTLS(tlsCrt, tlsKey))

}

func base64FileToStringFile(file string) (string, error) {

	fileByte, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	var decodeFileBytefile []byte
	_, err1 := base64.RawStdEncoding.Decode(decodeFileBytefile, fileByte)
	if err1 != nil {
		log.Printf("file %s base64 decode failed, use it directly", file)
		return file, err
	}
	dir := filepath.Dir(file)
	decodefilePath := dir + "/" + "decoded-" + file

	if err := os.WriteFile(decodefilePath, decodeFileBytefile, 0644); err != nil {
		log.Fatal("write file error: ", err)
		return decodefilePath, err
	}
	return decodefilePath, nil
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
