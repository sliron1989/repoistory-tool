package templates

import "fmt"

func GoMod(d RepoData) string {
	return fmt.Sprintf(`module github.com/%s/%s

go 1.26
`, d.Owner, d.Name)
}

func Microservice(d RepoData) string {
	return fmt.Sprintf(`package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var startTime = time.Now()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/health", handleHealth)

	log.Printf("%s starting on port %%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"service": "%s",
		"message": "Hello from %s!",
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"uptime": fmt.Sprintf("%%s", time.Since(startTime).Round(time.Second)),
	})
}
`, d.Name, d.Name, d.Name)
}
