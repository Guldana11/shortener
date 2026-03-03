package grpcserver

import (
	"context"

	"github.com/Guldana11/shortener/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ctxKey string

const userIDKey ctxKey = "userID"

// AuthInterceptor extracts userID from the "authorization" metadata header.
// It does NOT reject unauthenticated requests — each RPC decides its own policy.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("authorization"); len(vals) > 0 {
				if userID, err := service.ValidateToken(vals[0]); err == nil {
					ctx = context.WithValue(ctx, userIDKey, userID)
				}
			}
		}
		return handler(ctx, req)
	}
}

// UserIDFromContext extracts the userID set by the auth interceptor.
func UserIDFromContext(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(userIDKey).(string)
	return uid, ok && uid != ""
}
