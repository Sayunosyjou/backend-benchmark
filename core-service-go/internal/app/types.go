package app

import "time"

type Post struct {
	PostID    string   `bson:"postId" json:"postId"`
	AuthorID  string   `bson:"authorId" json:"authorId"`
	Content   string   `bson:"content" json:"content"`
	Tags      []string `bson:"tags" json:"tags"`
	CreatedAt int64    `bson:"createdAt" json:"createdAt"`
	LikeCount int64    `bson:"likeCount" json:"likeCount"`
	Status    string   `bson:"status" json:"status"`
}

func NowMillis() int64 { return time.Now().UnixMilli() }
