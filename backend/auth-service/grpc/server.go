package grpc

import (
	"auth-service/proto"
	"context"

	"github.com/golang-jwt/jwt"
)

type AuthServer struct {
	proto.UnimplementedAuthServiceServer
	jwtSecret string
}

func NewAuthServer(jwtSecret string) *AuthServer {
	return &AuthServer{jwtSecret: jwtSecret}
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *proto.TokenRequest) (*proto.TokenResponse, error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return &proto.TokenResponse{Valid: false}, nil
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)

	return &proto.TokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}
