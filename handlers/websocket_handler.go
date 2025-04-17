package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"video-ranking/config"
	"video-ranking/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ctx               = config.GetContext()
	rdb               = config.GetRedisClient()
	videoCollection   = config.GetVideoCollection()
	clientsAllVideos  = config.GetClientsAllVideos()
	clientsUserVideos = config.GetClientsUserVideos()
)

func getAllVideos() ([]map[string]interface{}, error) {
	videos, err := rdb.ZRevRangeWithScores(ctx, "global_ranking", 0, -1).Result()
	if err != nil {
		log.Println("Redis fetch error:", err)
		return nil, err
	}
	log.Printf("Fetched %d videos from Redis", len(videos))

	videoIDs := make([]string, len(videos))
	for i, v := range videos {
		videoIDs[i] = v.Member.(string)
	}

	filter := bson.M{"video_id": bson.M{"$in": videoIDs}}
	cursor, err := videoCollection.Find(ctx, filter)
	if err != nil {
		log.Println("MongoDB find error:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	videoList := make([]map[string]interface{}, 0, len(videos))
	videoMap := make(map[string]models.Video)
	for cursor.Next(ctx) {
		var video models.Video
		if err := cursor.Decode(&video); err != nil {
			log.Println("Mongo decode error:", err)
			continue
		}
		videoMap[video.VideoID] = video
	}
	log.Printf("Fetched %d videos from MongoDB", len(videoMap))

	for _, v := range videos {
		videoID := v.Member.(string)
		if video, exists := videoMap[videoID]; exists {
			videoList = append(videoList, map[string]interface{}{
				"video_id": videoID,
				"score":    v.Score,
				"user_id":  video.UserID,
			})
		}
	}
	log.Printf("Returning %d videos in videoList", len(videoList))
	return videoList, nil
}

func getUserVideos(userID string) ([]map[string]interface{}, error) {
	videos, err := rdb.ZRevRangeWithScores(ctx, "global_ranking", 0, -1).Result()
	if err != nil {
		log.Println("Redis fetch error:", err)
		return nil, err
	}
	log.Printf("Fetched %d videos from Redis for user %s", len(videos), userID)

	videoIDs := make([]string, len(videos))
	for i, v := range videos {
		videoIDs[i] = v.Member.(string)
	}

	filter := bson.M{
		"video_id": bson.M{"$in": videoIDs},
		"user_id":  userID,
	}
	cursor, err := videoCollection.Find(ctx, filter)
	if err != nil {
		log.Println("MongoDB find error for user", userID, ":", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	videoList := make([]map[string]interface{}, 0)
	videoMap := make(map[string]float64)
	for _, v := range videos {
		videoMap[v.Member.(string)] = v.Score
	}

	for cursor.Next(ctx) {
		var video models.Video
		if err := cursor.Decode(&video); err != nil {
			log.Println("Mongo decode error:", err)
			continue
		}
		if score, exists := videoMap[video.VideoID]; exists {
			videoList = append(videoList, map[string]interface{}{
				"video_id": video.VideoID,
				"score":    score,
			})
		}
	}
	log.Printf("Returning %d videos for user %s", len(videoList), userID)
	return videoList, nil
}

func HandleWebSocketVideos(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	clientsAllVideos[conn] = true
	defer func() {
		delete(clientsAllVideos, conn)
		conn.Close()
	}()

	videoList, err := getAllVideos()
	if err != nil {
		log.Println("Error fetching all videos:", err)
		return
	}
	msg, _ := json.Marshal(videoList)
	log.Println("Sending initial all videos to new client")
	err = conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("WebSocket write error:", err)
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func HandleWebSocketUserVideos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	if userID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	clientsUserVideos[conn] = userID
	defer func() {
		delete(clientsUserVideos, conn)
		conn.Close()
	}()

	videoList, err := getUserVideos(userID)
	if err != nil {
		log.Println("Error fetching user videos:", err)
		return
	}
	msg, _ := json.Marshal(videoList)
	log.Printf("Sending initial user videos for %s to new client", userID)
	err = conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("WebSocket write error:", err)
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
