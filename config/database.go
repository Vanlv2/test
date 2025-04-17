package config

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	rdb               *redis.Client
	ctx               = context.Background()
	videoCollection   *mongo.Collection
	clientsAllVideos  = make(map[*websocket.Conn]bool)
	clientsUserVideos = make(map[*websocket.Conn]string)
	updateChan        = make(chan string, 100) // Channel gửi user_id của video thay đổi
)

func init() {
	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	mongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(mongoCtx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	videoCollection = client.Database("video_db").Collection("videos")
}

func GetRedisClient() *redis.Client {
	return rdb
}

func GetContext() context.Context {
	return ctx
}

func GetVideoCollection() *mongo.Collection {
	return videoCollection
}

func GetClientsAllVideos() map[*websocket.Conn]bool {
	return clientsAllVideos
}

func GetClientsUserVideos() map[*websocket.Conn]string {
	return clientsUserVideos
}

func GetUpdateChan() chan string {
	return updateChan
}
