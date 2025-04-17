package main

import (
	"log"
	"net/http"

	"video-ranking/handlers"

	"github.com/gorilla/mux"
)

func main() {
	go handlers.KafkaConsume()
	go handlers.HandleMessages()

	r := mux.NewRouter()
	r.HandleFunc("/ws/videos", handlers.HandleWebSocketVideos)
	r.HandleFunc("/ws/user/{user_id}", handlers.HandleWebSocketUserVideos)
	r.HandleFunc("/interact", handlers.HandleInteraction).Methods("POST")

	log.Println("Server running at :8080")
	http.ListenAndServe(":8080", r)
}
