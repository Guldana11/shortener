package grpcserver

import (
	"context"
	"errors"

	pb "github.com/Guldana11/shortener/api/shortener"
	"github.com/Guldana11/shortener/internal/model"
	"github.com/Guldana11/shortener/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// URLShortener defines the service interface used by the gRPC layer.
type URLShortener interface {
	ShortenURL(userID, originalURL string) (service.ShortenResult, error)
	ExpandURL(id string) (string, error)
	ListUserURLs(userID string) []service.URLPair
}

// ShortenerServer implements the gRPC ShortenerService.
type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	Svc URLShortener
}

func (s *ShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "url is required")
	}

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		var token string
		userID, token = service.GenerateToken()
		md := metadata.Pairs("authorization", token)
		_ = grpc.SendHeader(ctx, md)
	}

	result, err := s.Svc.ShortenURL(userID, req.GetUrl())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	if result.Conflict {
		st, _ := status.New(codes.AlreadyExists, "url already exists").
			WithDetails(&pb.URLShortenResponse{Result: result.ShortURL})
		return nil, st.Err()
	}

	return &pb.URLShortenResponse{Result: result.ShortURL}, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	original, err := s.Svc.ExpandURL(req.GetId())
	if err != nil {
		if errors.Is(err, model.ErrNotFound) || errors.Is(err, model.ErrDeleted) {
			return nil, status.Error(codes.NotFound, "url not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.URLExpandResponse{Result: original}, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authorization required")
	}

	urls := s.Svc.ListUserURLs(userID)

	resp := &pb.UserURLsResponse{}
	for _, u := range urls {
		resp.Url = append(resp.Url, &pb.URLData{
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
		})
	}

	return resp, nil
}
