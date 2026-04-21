package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type createReq struct {
	AuthorID string   `json:"authorId"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
}

type createResp struct {
	PostID string `json:"postId"`
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func token(secret, user string) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": user, "exp": time.Now().Add(5 * time.Minute).Unix()})
	s, _ := t.SignedString([]byte(secret))
	return s
}

func randHex() string { b := make([]byte, 4); _, _ = rand.Read(b); return hex.EncodeToString(b) }

func main() {
	base := getenv("GATEWAY_BASE", "http://gateway:8088")
	count, _ := strconv.Atoi(getenv("SEED_POST_COUNT", "5000"))
	users, _ := strconv.Atoi(getenv("SEED_USER_COUNT", "200"))
	secret := getenv("JWT_SECRET", "dev-secret")
	out := getenv("OUTPUT_FILE", "/out/post_ids.txt")
	if err := os.MkdirAll("/out", 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; i < count; i++ {
		uid := fmt.Sprintf("u-%03d", i%users)
		reqBody, _ := json.Marshal(createReq{AuthorID: uid, Content: fmt.Sprintf("seed post %d %s", i, randHex()), Tags: []string{"tech", "bench"}})
		req, _ := http.NewRequest("POST", base+"/api/v1/posts", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token(secret, uid))
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		var c createResp
		_ = json.NewDecoder(resp.Body).Decode(&c)
		_ = resp.Body.Close()
		if c.PostID != "" {
			_, _ = fmt.Fprintln(f, c.PostID)
			if i%5 == 0 {
				r, _ := http.NewRequest("POST", base+"/api/v1/posts/"+c.PostID+"/like", nil)
				r.Header.Set("Authorization", "Bearer "+token(secret, uid))
				_, _ = client.Do(r)
			}
		}
		if i%200 == 0 {
			log.Printf("seeded %d/%d", i, count)
		}
	}
	log.Printf("seed finished: %d", count)
}
