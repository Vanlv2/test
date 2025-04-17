package models

type Interaction struct {
	VideoID   string  `json:"video_id" bson:"video_id"`
	UserID    string  `json:"user_id" bson:"user_id"`
	Type      string  `json:"type" bson:"type"`
	WatchTime float64 `json:"watch_time,omitempty" bson:"watch_time,omitempty"`
	Timestamp int64   `json:"timestamp" bson:"timestamp"`
}
