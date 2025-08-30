package service

import (
	"context"
	"time"

	"github.com/ahmdfkhri/hydrocast/backend/internal/model"
	"github.com/ahmdfkhri/hydrocast/backend/internal/repository"
	"github.com/ahmdfkhri/hydrocast/backend/internal/types"
	"github.com/ahmdfkhri/hydrocast/backend/pkg/auth"
	"github.com/ahmdfkhri/hydrocast/backend/pkg/pb"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthServer struct {
	pb.UnimplementedAuthServer
	userRepo   *repository.UserRepository
	authorizer *auth.Authorizer
}

func NewAuthServer(userRepo *repository.UserRepository, authorizer *auth.Authorizer) *AuthServer {
	return &AuthServer{
		userRepo:   userRepo,
		authorizer: authorizer,
	}
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Get user by username or email
	user, err := s.userRepo.GetByUsernameOrEmail(ctx, req.UsernameOrEmail)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	// Generate tokens
	accessToken, err := s.authorizer.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}

	refreshToken, err := s.authorizer.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServer) Signup(ctx context.Context, req *pb.SignupRequest) (*pb.SignupResponse, error) {
	// check if username or email are already used
	existingUser, _ := s.userRepo.GetByUsernameOrEmail(ctx, req.Username)
	if existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "username already exists")
	}

	existingUser, _ = s.userRepo.GetByUsernameOrEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, status.Error(codes.AlreadyExists, "email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     types.UR_User,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// Generate access token
	accessToken, err := s.authorizer.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
	}

	// Generate refresh token
	refreshToken, err := s.authorizer.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
	}

	return &pb.SignupResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	// Verify the refresh token
	claims, err := s.authorizer.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	// Get user to ensure they still exist and haven't been deleted/disabled
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Generate new access token
	accessToken, err := s.authorizer.GenerateToken(user.ID, user.Role, types.TT_Access)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}

	var refreshTokenPtr *string

	// Check if refresh token needs to be rotated (re-issued)
	if claims.IssuedAt != nil {
		timeSinceIssued := time.Since(claims.IssuedAt.Time)
		if timeSinceIssued > s.authorizer.ReRefreshDuration {
			// Generate new refresh token since it's been longer than reRefreshDuration
			newRefreshToken, err := s.authorizer.GenerateToken(user.ID, user.Role, types.TT_Refresh)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to generate refresh token")
			}
			refreshTokenPtr = &newRefreshToken
		}
		// If not exceeded reRefreshDuration, don't generate new refresh token (refreshTokenPtr remains nil)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenPtr, // Will be nil if not regenerated
	}, nil
}

func (s *AuthServer) GetProfile(ctx context.Context, _ *emptypb.Empty) (*pb.GetProfileResponse, error) {
	// Extract user ID from gRPC metadata via helper
	userID, err := auth.GetUserID(ctx)
	if err != nil {
		return nil, err // already returns appropriate gRPC status
	}

	// Fetch user from repository
	user, err := s.userRepo.GetByID(ctx, *userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Return user profile
	return &pb.GetProfileResponse{
		Id:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
	}, nil
}
