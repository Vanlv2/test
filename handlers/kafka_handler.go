package handlers

import (
	"encoding/json"
	"log"

	"video-ranking/models"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func calculateScore(interaction models.Interaction) float64 {
	switch interaction.Type {
	case "view":
		return 1
	case "like":
		return 2
	case "comment":
		return 2
	case "share":
		return 3
	case "watch_time":
		if interaction.WatchTime < 2.0 {
			return 1
		} else {
			return 2
		}
	default:
		return 0
	}
}

func updateVideoInMongo(videoID string, interaction models.Interaction) {
	filter := bson.M{"video_id": videoID}
	var video models.Video
	err := videoCollection.FindOne(ctx, filter).Decode(&video)
	score := calculateScore(interaction)

	if err == mongo.ErrNoDocuments {
		newVideo := models.Video{
			VideoID:        videoID,
			UserID:         interaction.UserID,
			Score:          score,
			Views:          1,
			Likes:          0,
			Comments:       0,
			Shares:         0,
			WatchTimeTotal: interaction.WatchTime,
			// WatchTimeScore: score,
		}
		switch interaction.Type {
		case "like":
			newVideo.Likes = 1
		case "comment":
			newVideo.Comments = 1
		case "share":
			newVideo.Shares = 1
		}
		_, err = videoCollection.InsertOne(ctx, newVideo)
		if err != nil {
			log.Println("Mongo insert error:", err)
		} else {
			log.Printf("Inserted new video: video_id=%s, user_id=%s", videoID, interaction.UserID)
		}
	} else if err == nil {
		update := bson.M{
			"$inc": bson.M{
				"views":            1,
				"likes":            0,
				"comments":         0,
				"shares":           0,
				"watch_time_total": interaction.WatchTime,
				// "watch_time_score": score,
			},
			"$set": bson.M{
				"score": video.Score + score,
			},
		}
		switch interaction.Type {
		case "like":
			update["$inc"].(bson.M)["likes"] = 1
		case "comment":
			update["$inc"].(bson.M)["comments"] = 1
		case "share":
			update["$inc"].(bson.M)["shares"] = 1
		}
		_, err = videoCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Println("Mongo update error:", err)
		} else {
			log.Printf("Updated video: video_id=%s, score=%f", videoID, video.Score+score)
		}
	} else {
		log.Println("Mongo find error:", err)
	}
}

func KafkaConsume() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "video_interactions",
		GroupID: "ranking_group",
	})
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Println("Kafka read error:", err)
			continue
		}
		var inter models.Interaction
		err = json.Unmarshal(m.Value, &inter)
		if err != nil {
			log.Println("Invalid message format:", err)
			continue
		}
		log.Printf("Received interaction: video_id=%s, type=%s", inter.VideoID, inter.Type)
		updateVideoInMongo(inter.VideoID, inter)
		score := calculateScore(inter)
		_, err = rdb.ZIncrBy(ctx, "global_ranking", score, inter.VideoID).Result()
		if err != nil {
			log.Println("Redis ZIncrBy error:", err)
		} else {
			log.Printf("Updated Redis: video_id=%s, score_increment=%f", inter.VideoID, score)
		}

		// Lấy user_id của video từ MongoDB
		var video models.Video
		err = videoCollection.FindOne(ctx, bson.M{"video_id": inter.VideoID}).Decode(&video)
		if err != nil {
			log.Println("Mongo find user_id error:", err)
			continue
		}

		// Thông báo sự kiện thay đổi với user_id
		select {
		case updateChan <- video.UserID:
			log.Printf("Notified update event for user_id=%s", video.UserID)
		default:
			log.Println("Update channel full, skipping notification")
		}
	}
}
