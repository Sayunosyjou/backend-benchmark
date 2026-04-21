package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Event struct {
	Type      string   `json:"type"`
	PostID    string   `json:"postId"`
	AuthorID  string   `json:"authorId,omitempty"`
	Content   string   `json:"content,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	LikeDelta int64    `json:"likeDelta,omitempty"`
	CreatedAt int64    `json:"createdAt"`
}

type Post struct {
	PostID    string   `bson:"postId"`
	AuthorID  string   `bson:"authorId"`
	Content   string   `bson:"content"`
	Tags      []string `bson:"tags"`
	CreatedAt int64    `bson:"createdAt"`
	LikeCount int64    `bson:"likeCount,omitempty"`
	Status    string   `bson:"status"`
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	ctx := context.Background()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(getenv("MONGO_URI", "mongodb://mongo:27017")))
	if err != nil {
		log.Fatal(err)
	}
	coll := mongoClient.Database(getenv("MONGO_DB", "social")).Collection(getenv("MONGO_COLLECTION", "posts"))
	rdb := redis.NewClient(&redis.Options{Addr: getenv("VALKEY_ADDR", "valkey:6379")})
	batchMs, _ := strconv.Atoi(getenv("BATCH_FLUSH_MS", "200"))
	maxBatch, _ := strconv.Atoi(getenv("MAX_BATCH_SIZE", "500"))

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_1_0_0
	cfg.Consumer.Return.Errors = true
	group, err := sarama.NewConsumerGroup(strings.Split(getenv("KAFKA_BROKER", "redpanda:9092"), ","), getenv("KAFKA_GROUP", "post-consumer"), cfg)
	if err != nil {
		log.Fatal(err)
	}

	h := &handler{coll: coll, rdb: rdb, flushEvery: time.Duration(batchMs) * time.Millisecond, maxBatch: maxBatch}
	for {
		if err := group.Consume(ctx, []string{getenv("KAFKA_TOPIC", "post-events")}, h); err != nil {
			log.Printf("consume: %v", err)
			time.Sleep(time.Second)
		}
	}
}

type handler struct {
	coll       *mongo.Collection
	rdb        *redis.Client
	flushEvery time.Duration
	maxBatch   int
	mu         sync.Mutex
	events     []Event
}

func (h *handler) Setup(sarama.ConsumerGroupSession) error   { go h.loopFlush(); return nil }
func (h *handler) Cleanup(sarama.ConsumerGroupSession) error { return h.flush(context.Background()) }
func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var e Event
		if err := json.Unmarshal(msg.Value, &e); err == nil {
			h.mu.Lock()
			h.events = append(h.events, e)
			full := len(h.events) >= h.maxBatch
			h.mu.Unlock()
			if full {
				_ = h.flush(sess.Context())
			}
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (h *handler) loopFlush() {
	t := time.NewTicker(h.flushEvery)
	defer t.Stop()
	for range t.C {
		_ = h.flush(context.Background())
	}
}

func (h *handler) flush(ctx context.Context) error {
	h.mu.Lock()
	batch := h.events
	h.events = nil
	h.mu.Unlock()
	if len(batch) == 0 {
		return nil
	}
	models := make([]mongo.WriteModel, 0, len(batch))
	for _, e := range batch {
		switch e.Type {
		case "post-created":
			models = append(models, mongo.NewUpdateOneModel().SetFilter(bson.M{"postId": e.PostID}).SetUpdate(bson.M{"$set": Post{PostID: e.PostID, AuthorID: e.AuthorID, Content: e.Content, Tags: e.Tags, CreatedAt: e.CreatedAt, Status: "READY"}, "$setOnInsert": bson.M{"likeCount": 0}}).SetUpsert(true))
			_ = h.rdb.HSet(ctx, "post:"+e.PostID, map[string]interface{}{"status": "READY", "postId": e.PostID, "authorId": e.AuthorID, "content": e.Content, "tags": strings.Join(e.Tags, ","), "createdAt": e.CreatedAt}).Err()
		case "post-liked":
			models = append(models, mongo.NewUpdateOneModel().SetFilter(bson.M{"postId": e.PostID}).SetUpdate(bson.M{"$inc": bson.M{"likeCount": e.LikeDelta}}).SetUpsert(true))
		}
	}
	_, err := h.coll.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		log.Printf("bulk write: %v", err)
	}
	return err
}
