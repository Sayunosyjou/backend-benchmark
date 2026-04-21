package com.example.web;

import com.example.social.v1.*;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Size;
import org.springframework.http.HttpStatus;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@Validated
public class PostController {
    private final PostServiceGrpc.PostServiceBlockingStub stub;

    public PostController(PostServiceGrpc.PostServiceBlockingStub stub) {
        this.stub = stub;
    }

    record CreatePostReq(@NotBlank String authorId, @NotBlank @Size(max = 1000) String content, List<String> tags) {}

    @PostMapping("/api/v1/posts")
    @ResponseStatus(HttpStatus.CREATED)
    public Map<String, Object> create(@RequestBody @Valid CreatePostReq req) {
        var resp = stub.createPost(CreatePostRequest.newBuilder().setAuthorId(req.authorId()).setContent(req.content()).addAllTags(req.tags() == null ? List.of() : req.tags()).build());
        return Map.of("postId", resp.getPostId(), "createdAt", resp.getCreatedAtUnixMs(), "status", resp.getStatus());
    }

    @GetMapping("/api/v1/posts/{postId}")
    public Map<String, Object> getPost(@PathVariable String postId) {
        var p = stub.getPost(GetPostRequest.newBuilder().setPostId(postId).build()).getPost();
        return Map.of("postId", p.getPostId(), "authorId", p.getAuthorId(), "content", p.getContent(), "tags", p.getTagsList(), "createdAt", p.getCreatedAtUnixMs(), "likeCount", p.getLikeCount(), "status", p.getStatus());
    }

    @GetMapping("/api/v1/feed/hot")
    public List<Map<String, Object>> hot(@RequestParam(defaultValue = "50") int limit) {
        return stub.getHotFeed(GetHotFeedRequest.newBuilder().setLimit(limit).build()).getPostsList().stream().map(p -> Map.<String, Object>of(
                "postId", p.getPostId(), "authorId", p.getAuthorId(), "content", p.getContent(), "tags", p.getTagsList(), "createdAt", p.getCreatedAtUnixMs(), "likeCount", p.getLikeCount(), "status", p.getStatus()
        )).toList();
    }

    @PostMapping("/api/v1/posts/{postId}/like")
    public Map<String, Object> like(@PathVariable String postId, @RequestHeader(value = "X-User-Id", defaultValue = "bench-user") String userId) {
        var r = stub.likePost(LikePostRequest.newBuilder().setPostId(postId).setUserId(userId).build());
        return Map.of("postId", r.getPostId(), "likeCount", r.getLikeCount());
    }

    @GetMapping("/healthz")
    public Map<String, String> health() { return Map.of("status", "ok"); }

    @GetMapping("/readyz")
    public Map<String, String> ready() {
        var h = stub.health(HealthRequest.newBuilder().build());
        return Map.of("status", h.getStatus());
    }
}
