package handlers

import (
	"encoding/json"
	"log"
	"video-ranking/config"

	"github.com/gorilla/websocket"
)

var (
	updateChan = config.GetUpdateChan()
)

func HandleMessages() {
	for {
		// Chờ user_id từ updateChan
		userID := <-updateChan
		log.Printf("Received update event for user_id=%s", userID)

		// Cập nhật danh sách tất cả video cho /ws/videos
		videoList, err := getAllVideos()
		if err != nil {
			log.Println("Error fetching all videos:", err)
			continue
		}
		msgAllVideos, _ := json.Marshal(videoList)
		log.Printf("Sending all videos to %d clients", len(clientsAllVideos))
		for client := range clientsAllVideos {
			err := client.WriteMessage(websocket.TextMessage, msgAllVideos)
			if err != nil {
				log.Println("WebSocket write error for all videos client:", err)
				client.Close()
				delete(clientsAllVideos, client)
			}
		}

		// Cập nhật danh sách video cho /ws/user/{user_id} chỉ khi user_id khớp
		log.Printf("Sending user videos to %d clients", len(clientsUserVideos))
		for client, clientUserID := range clientsUserVideos {
			if clientUserID == userID {
				videoList, err := getUserVideos(clientUserID)
				if err != nil {
					log.Println("Error fetching user videos for", clientUserID, ":", err)
					continue
				}
				msg, _ := json.Marshal(videoList)
				err = client.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					log.Println("WebSocket write error for user", clientUserID, ":", err)
					client.Close()
					delete(clientsUserVideos, client)
				}
			}
		}
	}
}
