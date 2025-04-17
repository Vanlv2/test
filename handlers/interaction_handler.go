package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"video-ranking/models"

	"github.com/segmentio/kafka-go"
)

func HandleInteraction(w http.ResponseWriter, r *http.Request) {
	var inter models.Interaction
	err := json.NewDecoder(r.Body).Decode(&inter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	inter.Timestamp = time.Now().Unix()
	msg, _ := json.Marshal(inter)
	writer := kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "video_interactions",
		Balancer: &kafka.LeastBytes{},
	}
	err = writer.WriteMessages(ctx, kafka.Message{
		Value: msg,
	})
	if err != nil {
		http.Error(w, "Kafka write failed", http.StatusInternalServerError)
		return
	}
	log.Println("Interaction sent to Kafka:", string(msg))
	w.Write([]byte("Interaction sent to Kafka"))
}
