package grpcserver

import (
	"context"
	"time"

	"core-service-go/internal/infra"
	pb "core-service-go/proto"
)

type Server struct {
	pb.UnimplementedPostServiceServer
	Store *infra.Store
}

func (s *Server) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {
	p, err := s.Store.CreatePost(ctx, req.GetAuthorId(), req.GetContent(), req.GetTags())
	if err != nil {
		return nil, err
	}
	return &pb.CreatePostResponse{PostId: p.PostID, CreatedAtUnixMs: p.CreatedAt, Status: p.Status}, nil
}

func (s *Server) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.GetPostResponse, error) {
	p, err := s.Store.GetPost(ctx, req.GetPostId())
	if err != nil {
		return nil, err
	}
	return &pb.GetPostResponse{Post: &pb.Post{PostId: p.PostID, AuthorId: p.AuthorID, Content: p.Content, Tags: p.Tags, CreatedAtUnixMs: p.CreatedAt, LikeCount: p.LikeCount, Status: p.Status}}, nil
}

func (s *Server) GetHotFeed(ctx context.Context, req *pb.GetHotFeedRequest) (*pb.GetHotFeedResponse, error) {
	if req.GetLimit() <= 0 {
		req.Limit = 50
	}
	posts, err := s.Store.GetHotFeed(ctx, int64(req.GetLimit()))
	if err != nil {
		return nil, err
	}
	out := make([]*pb.Post, 0, len(posts))
	for _, p := range posts {
		out = append(out, &pb.Post{PostId: p.PostID, AuthorId: p.AuthorID, Content: p.Content, Tags: p.Tags, CreatedAtUnixMs: p.CreatedAt, LikeCount: p.LikeCount, Status: p.Status})
	}
	return &pb.GetHotFeedResponse{Posts: out}, nil
}

func (s *Server) LikePost(ctx context.Context, req *pb.LikePostRequest) (*pb.LikePostResponse, error) {
	cnt, err := s.Store.LikePost(ctx, req.GetPostId(), req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &pb.LikePostResponse{PostId: req.GetPostId(), LikeCount: cnt}, nil
}

func (s *Server) Health(ctx context.Context, _ *pb.HealthRequest) (*pb.HealthResponse, error) {
	status := "UP"
	if err := s.Store.Health(ctx); err != nil {
		status = "DEGRADED"
	}
	return &pb.HealthResponse{Status: status, ServerTimeUnixMs: time.Now().UnixMilli()}, nil
}
