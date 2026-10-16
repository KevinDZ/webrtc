package main

import (
	"log"
	"net/http"

	"webrtc/internal/agent"
	"webrtc/internal/signaling"
)

func main() {
	svc := agent.NewService()
	handler := signaling.NewServer(svc)

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!\n"))
	})
	http.HandleFunc("/ws", handler.ServeHTTP)

	log.Println("agent signaling listening on :8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
