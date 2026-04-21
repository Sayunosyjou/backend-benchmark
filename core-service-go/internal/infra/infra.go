package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"core-service-go/internal/app"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	Redis          *redis.Client
	Mongo          *mongo.Collection
	Producer       sarama.SyncProducer
	Topic          string
	HotKey         string
	RecentKey      string
	MaxCacheRecent int64
}

type Event struct {
	Type      string   `json:"type"`
	PostID    string   `json:"postId"`
	AuthorID  string   `json:"authorId,omitempty"`
	Content   string   `json:"content,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	LikeDelta int64    `json:"likeDelta,omitempty"`
	CreatedAt int64    `json:"createdAt"`
}

func NewStore(ctx context.Context, redisAddr, mongoURI, mongoDB, mongoColl, broker, topic string, maxRecent int64) (*Store, error) {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	mc, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	if err = mc.Ping(ctx, nil); err != nil {
		return nil, err
	}
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_1_0_0
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	producer, err := sarama.NewSyncProducer([]string{broker}, cfg)
	if err != nil {
		return nil, err
	}
	return &Store{Redis: rdb, Mongo: mc.Database(mongoDB).Collection(mongoColl), Producer: producer, Topic: topic, HotKey: "hot:posts", RecentKey: "recent:posts", MaxCacheRecent: maxRecent}, nil
}

func (s *Store) CreatePost(ctx context.Context, authorID, content string, tags []string) (*app.Post, error) {
	id := uuid.NewString()
	post := &app.Post{PostID: id, AuthorID: authorID, Content: content, Tags: tags, CreatedAt: app.NowMillis(), LikeCount: 0, Status: "PENDING"}
	if err := s.cachePost(ctx, post); err != nil {
		return nil, err
	}
	pipe := s.Redis.Pipeline()
	pipe.ZAdd(ctx, s.HotKey, redis.Z{Score: float64(post.CreatedAt), Member: post.PostID})
	pipe.LPush(ctx, s.RecentKey, post.PostID)
	pipe.LTrim(ctx, s.RecentKey, 0, s.MaxCacheRecent-1)
	_, _ = pipe.Exec(ctx)
	_ = s.Publish(ctx, Event{Type: "post-created", PostID: post.PostID, AuthorID: authorID, Content: content, Tags: tags, CreatedAt: post.CreatedAt})
	return post, nil
}

func (s *Store) GetPost(ctx context.Context, postID string) (*app.Post, error) {
	post, err := s.getCachedPost(ctx, postID)
	if err == nil && post != nil {
		if post.Status != "READY" {
			return nil, mongo.ErrNoDocuments
		}
		return post, nil
	}
	var m app.Post
	if err := s.Mongo.FindOne(ctx, bson.M{"postId": postID}).Decode(&m); err != nil {
		return nil, err
	}
	m.Status = "READY"
	_ = s.cachePost(ctx, &m)
	return &m, nil
}

func (s *Store) GetHotFeed(ctx context.Context, limit int64) ([]*app.Post, error) {
	ids, err := s.Redis.ZRevRange(ctx, s.HotKey, 0, limit-1).Result()
	if err != nil {
		ids = nil
	}
	posts := make([]*app.Post, 0, limit)
	for _, id := range ids {
		p, err := s.GetPost(ctx, id)
		if err == nil && p != nil {
			posts = append(posts, p)
		}
	}
	if len(posts) > 0 {
		return posts, nil
	}
	cur, err := s.Mongo.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var p app.Post
		if err := cur.Decode(&p); err == nil {
			p.Status = "READY"
			_ = s.cachePost(ctx, &p)
			_ = s.Redis.ZAdd(ctx, s.HotKey, redis.Z{Score: float64(p.LikeCount) + float64(p.CreatedAt), Member: p.PostID}).Err()
			posts = append(posts, &p)
		}
	}
	return posts, nil
}

func (s *Store) LikePost(ctx context.Context, postID, userID string) (int64, error) {
	likesKey := "likes:" + postID
	newLikes, err := s.Redis.Incr(ctx, likesKey).Result()
	if err != nil {
		return 0, err
	}
	_ = s.Redis.HSet(ctx, "post:"+postID, "likeCount", newLikes).Err()
	_ = s.Redis.ZIncrBy(ctx, s.HotKey, 1, postID).Err()
	_ = s.Publish(ctx, Event{Type: "post-liked", PostID: postID, LikeDelta: 1, CreatedAt: app.NowMillis()})
	_ = userID
	return newLikes, nil
}

func (s *Store) Publish(ctx context.Context, event Event) error {
	b, _ := json.Marshal(event)
	_, _, err := s.Producer.SendMessage(&sarama.ProducerMessage{Topic: s.Topic, Value: sarama.ByteEncoder(b), Timestamp: time.Now()})
	if err != nil {
		log.Printf("publish error: %v", err)
	}
	return err
}

func (s *Store) cachePost(ctx context.Context, p *app.Post) error {
	payload := map[string]interface{}{
		"postId": p.PostID, "authorId": p.AuthorID, "content": p.Content,
		"tags": strings.Join(p.Tags, ","), "createdAt": p.CreatedAt,
		"likeCount": p.LikeCount, "status": p.Status,
	}
	return s.Redis.HSet(ctx, "post:"+p.PostID, payload).Err()
}

func (s *Store) getCachedPost(ctx context.Context, postID string) (*app.Post, error) {
	m, err := s.Redis.HGetAll(ctx, "post:"+postID).Result()
	if err != nil || len(m) == 0 {
		return nil, err
	}
	createdAt, _ := strconv.ParseInt(m["createdAt"], 10, 64)
	likeCount, _ := strconv.ParseInt(m["likeCount"], 10, 64)
	tags := []string{}
	if t := m["tags"]; t != "" {
		tags = strings.Split(t, ",")
	}
	return &app.Post{PostID: m["postId"], AuthorID: m["authorId"], Content: m["content"], Tags: tags, CreatedAt: createdAt, LikeCount: likeCount, Status: m["status"]}, nil
}

func (s *Store) Health(ctx context.Context) error {
	if err := s.Redis.Ping(ctx).Err(); err != nil {
		return err
	}
	return s.Mongo.Database().Client().Ping(ctx, nil)
}
