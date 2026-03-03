package grpcserver

import (
	"context"
	"net"
	"testing"

	pb "github.com/Guldana11/shortener/api/shortener"
	"github.com/Guldana11/shortener/internal/audit"
	"github.com/Guldana11/shortener/internal/repository"
	"github.com/Guldana11/shortener/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func startTestServer(t *testing.T) (pb.ShortenerServiceClient, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	repo := repository.NewURLRepository("")
	logger, _ := zap.NewDevelopment()
	publisher := audit.NewPublisher()
	svc := &service.URLService{
		Repo:      repo,
		BaseURL:   "http://localhost:8080",
		Publisher: publisher,
		Logger:    logger,
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(AuthInterceptor()))
	pb.RegisterShortenerServiceServer(srv, &ShortenerServer{Svc: svc})

	go func() {
		_ = srv.Serve(lis)
	}()

	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	client := pb.NewShortenerServiceClient(conn)
	cleanup := func() {
		conn.Close()
		srv.GracefulStop()
	}
	return client, cleanup
}

func TestShortenURL_Success(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	resp, err := client.ShortenURL(context.Background(), &pb.URLShortenRequest{Url: "https://example.com"})
	if err != nil {
		t.Fatalf("ShortenURL failed: %v", err)
	}
	if resp.GetResult() == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestShortenURL_EmptyURL(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	_, err := client.ShortenURL(context.Background(), &pb.URLShortenRequest{Url: ""})
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
	if s, ok := status.FromError(err); !ok || s.Code() != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestShortenURL_AlreadyExists(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	// Create first and capture the token
	var header metadata.MD
	resp1, err := client.ShortenURL(
		context.Background(),
		&pb.URLShortenRequest{Url: "https://duplicate.com"},
		grpc.Header(&header),
	)
	if err != nil {
		t.Fatalf("first ShortenURL failed: %v", err)
	}

	tokens := header.Get("authorization")
	if len(tokens) == 0 {
		t.Fatal("expected authorization token")
	}

	// Create again with same user — should get AlreadyExists with result in details
	md := metadata.Pairs("authorization", tokens[0])
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	_, err = client.ShortenURL(ctx, &pb.URLShortenRequest{Url: "https://duplicate.com"})
	if err == nil {
		t.Fatal("expected error for duplicate URL")
	}
	s, ok := status.FromError(err)
	if !ok || s.Code() != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists, got %v", err)
	}
	details := s.Details()
	if len(details) == 0 {
		t.Fatal("expected status details with result")
	}
	detail, ok := details[0].(*pb.URLShortenResponse)
	if !ok {
		t.Fatalf("expected URLShortenResponse detail, got %T", details[0])
	}
	if detail.GetResult() != resp1.GetResult() {
		t.Fatalf("expected same result for duplicate, got %s vs %s", detail.GetResult(), resp1.GetResult())
	}
}

func TestShortenURL_NoAuth_GeneratesToken(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	var header metadata.MD
	_, err := client.ShortenURL(
		context.Background(),
		&pb.URLShortenRequest{Url: "https://token-test.com"},
		grpc.Header(&header),
	)
	if err != nil {
		t.Fatalf("ShortenURL failed: %v", err)
	}

	tokens := header.Get("authorization")
	if len(tokens) == 0 {
		t.Fatal("expected authorization header in response metadata")
	}
	// Validate the token
	userID, err := service.ValidateToken(tokens[0])
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}
	if userID == "" {
		t.Fatal("expected non-empty userID")
	}
}

func TestExpandURL_Success(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	// Shorten first, extract ID
	var header metadata.MD
	resp, err := client.ShortenURL(
		context.Background(),
		&pb.URLShortenRequest{Url: "https://expand-test.com"},
		grpc.Header(&header),
	)
	if err != nil {
		t.Fatalf("ShortenURL failed: %v", err)
	}

	// Extract ID from result (last path segment)
	result := resp.GetResult()
	// result is like "http://localhost:8080/abc123"
	id := result[len("http://localhost:8080/"):]

	expandResp, err := client.ExpandURL(context.Background(), &pb.URLExpandRequest{Id: id})
	if err != nil {
		t.Fatalf("ExpandURL failed: %v", err)
	}
	if expandResp.GetResult() != "https://expand-test.com" {
		t.Fatalf("expected https://expand-test.com, got %s", expandResp.GetResult())
	}
}

func TestExpandURL_NotFound(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	_, err := client.ExpandURL(context.Background(), &pb.URLExpandRequest{Id: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for non-existent URL")
	}
	s, ok := status.FromError(err)
	if !ok || s.Code() != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestListUserURLs_Unauthenticated(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	_, err := client.ListUserURLs(context.Background(), &emptypb.Empty{})
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}
	s, ok := status.FromError(err)
	if !ok || s.Code() != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated, got %v", err)
	}
}

func TestListUserURLs_Empty(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	// Generate a valid token
	_, token := service.GenerateToken()
	md := metadata.Pairs("authorization", token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	resp, err := client.ListUserURLs(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("ListUserURLs failed: %v", err)
	}
	if len(resp.GetUrl()) != 0 {
		t.Fatalf("expected empty list, got %d items", len(resp.GetUrl()))
	}
}

func TestListUserURLs_WithURLs(t *testing.T) {
	client, cleanup := startTestServer(t)
	defer cleanup()

	// Create a URL and capture the token
	var header metadata.MD
	_, err := client.ShortenURL(
		context.Background(),
		&pb.URLShortenRequest{Url: "https://list-test.com"},
		grpc.Header(&header),
	)
	if err != nil {
		t.Fatalf("ShortenURL failed: %v", err)
	}

	tokens := header.Get("authorization")
	if len(tokens) == 0 {
		t.Fatal("expected authorization token")
	}

	// List with the same token
	md := metadata.Pairs("authorization", tokens[0])
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	resp, err := client.ListUserURLs(ctx, &emptypb.Empty{})
	if err != nil {
		t.Fatalf("ListUserURLs failed: %v", err)
	}
	if len(resp.GetUrl()) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(resp.GetUrl()))
	}
	if resp.GetUrl()[0].GetOriginalUrl() != "https://list-test.com" {
		t.Fatalf("expected https://list-test.com, got %s", resp.GetUrl()[0].GetOriginalUrl())
	}
}
