package grpc

import (
	"context"
	"time"

	"github.com/korneevDev/auth-service/internal/repository"
	"github.com/korneevDev/auth-service/pkg/jwt"

	pb "github.com/korneevDev/auth-service/proto/auth"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	userRepo           repository.UserRepository
	jwtSecret          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewAuthServer(
	repo repository.UserRepository,
	jwtSecret string,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
) *AuthServer {
	return &AuthServer{
		userRepo:           repo,
		jwtSecret:          jwtSecret,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	claims, err := jwt.ValidateToken(req.AccessToken, s.jwtSecret)
	if err != nil {
		return &pb.TokenResponse{Valid: false}, nil
	}

	return &pb.TokenResponse{
		Valid:  true,
		UserId: claims["sub"].(string),
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	user, err := s.userRepo.GetUserByRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	accessToken, _ := jwt.GenerateAccessToken(user, s.accessTokenExpiry, s.jwtSecret)
	refreshToken, _ := jwt.GenerateRefreshToken(user, s.refreshTokenExpiry, s.jwtSecret)

	s.userRepo.SaveRefreshToken(user.ID, refreshToken)

	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
