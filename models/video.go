package models

type Video struct {
	VideoID        string  `bson:"video_id" json:"video_id"`
	UserID         string  `bson:"user_id" json:"user_id"`
	Score          float64 `bson:"score" json:"score"`
	Views          int     `bson:"views" json:"views"`
	Likes          int     `bson:"likes" json:"likes"`
	Comments       int     `bson:"comments" json:"comments"`
	Shares         int     `bson:"shares" json:"shares"`
	WatchTimeTotal float64 `bson:"watch_time_total" json:"watch_time_total"`
}
