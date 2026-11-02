package service

import (
	"context"
	"errors"

	"github.com/Nurlan270/task-tracker-go/common/hasher"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"

	errs "github.com/Nurlan270/task-tracker-go/auth-service/internal/errors"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/grpc/dto/request"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/grpc/message"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/jwt"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/models"

	ctxlib "github.com/Nurlan270/task-tracker-go/common/context"
	authV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/auth/v1"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (uint, error)
	FindByEmail(ctx context.Context, email string) (models.User, error)
}

type TokenRepository interface {
	CreateRefresh(ctx context.Context, token *models.RefreshToken) error
	MarkRefreshAsRevokedByHash(ctx context.Context, hash string) error
	GetRefreshByHashWithUser(ctx context.Context, hash string) (models.RefreshToken, error)
}

type AuthService struct {
	authV1.UnimplementedAuthServiceServer

	tokenManager *jwt.TokenManager
	hasher       hasher.Hasher
	db           *gorm.DB
	userRepo     UserRepository
	tokenRepo    TokenRepository
}

func NewAuthService(
	tokenManager *jwt.TokenManager,
	hasher hasher.Hasher,
	db *gorm.DB,
	userRepo UserRepository,
	tokenRepo TokenRepository,
) *AuthService {
	return &AuthService{
		tokenManager: tokenManager,
		hasher:       hasher,
		db:           db,
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	req *authV1.RegisterRequest,
) (*authV1.RegisterResponse, error) {
	log := logger.Get().SetOp("Register")

	body := request.RegisterUser{
		Name:     req.GetName(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	//	Normalize request body
	body.Normalize()

	//	Validate request body
	if err := body.Validate(); err != nil {
		if !errs.IsValidation(err) {
			log.Error("internal validation error", zap.Error(err))

			return nil, status.Error(codes.Internal, message.ErrInternal)
		}

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	//	Hash password
	passHash, err := s.hasher.Hash(body.Password)
	if err != nil {
		log.Error("failed to hash password", zap.Error(err))

		return nil, status.Error(codes.Internal, message.ErrInternal)
	}

	var tokenPair jwt.TokenPair

	err = s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := ctxlib.NewTXCtx(ctx, tx)

		user := &models.User{
			Name:     body.Name,
			Email:    body.Email,
			Password: passHash,
		}

		//	Insert user into DB
		uid, err := s.userRepo.Create(txCtx, user)
		if err != nil {
			if errors.Is(err, errs.ErrUserAlreadyExists) {
				return status.Error(codes.AlreadyExists, message.ErrUserAlreadyExists)
			}

			log.Error("failed to create user", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		}

		claims := jwt.Claims{
			"user_id": uid,
			"name":    user.Name,
			"email":   user.Email,
		}

		//	Generate JWT token pair
		tokenPair, err = s.tokenManager.Generate(claims)
		if err != nil {
			log.Error("failed to generate jwt token pair", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		}

		refresh := &models.RefreshToken{
			UserID:    uid,
			TokenHash: tokenPair.RefreshToken.Hash,
			ExpiresAt: tokenPair.RefreshToken.ExpiresAt,
		}

		//	Insert refresh token into DB
		if err = s.tokenRepo.CreateRefresh(txCtx, refresh); err != nil {
			log.Error("failed to create refresh token", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		}

		return nil
	})
	if err != nil {
		//	We don't log here as we logged it already above
		return nil, err
	}

	return &authV1.RegisterResponse{
		AccessToken:  tokenPair.AccessToken.Value,
		RefreshToken: tokenPair.RefreshToken.Value,
		ExpiresIn:    int64(tokenPair.AccessToken.ExpiresIn.Seconds()),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *authV1.LoginRequest) (*authV1.LoginResponse, error) {
	log := logger.Get().SetOp("Login")

	body := request.LoginUser{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	//	Normalize request body
	body.Normalize()

	//	Validate request body
	if err := body.Validate(); err != nil {
		if !errs.IsValidation(err) {
			log.Error("internal validation error", zap.Error(err))

			return nil, status.Error(codes.Internal, message.ErrInternal)
		}

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var tokenPair jwt.TokenPair

	err := s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := ctxlib.NewTXCtx(ctx, tx)

		//	Check whether a user with email exists
		user, err := s.userRepo.FindByEmail(txCtx, body.Email)
		if err != nil {
			if errors.Is(err, errs.ErrUserNotFound) {
				return status.Error(codes.NotFound, message.ErrUserNotFound)
			}

			log.Error("failed to find user by email", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		}

		//	Verify password hashes
		match, err := s.hasher.Verify(body.Password, user.Password)
		if err != nil {
			log.Error("failed to verify password", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		} else if !match {
			return status.Error(codes.Unauthenticated, message.ErrUserNotFound)
		}

		claims := jwt.Claims{
			"user_id": user.ID,
			"name":    user.Name,
			"email":   user.Email,
		}

		//	Generate JWT token pair
		tokenPair, err = s.tokenManager.Generate(claims)
		if err != nil {
			log.Error("failed to generate jwt token pair", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		}

		refresh := &models.RefreshToken{
			UserID:    user.ID,
			TokenHash: tokenPair.RefreshToken.Hash,
			ExpiresAt: tokenPair.RefreshToken.ExpiresAt,
		}

		//	Insert refresh token into DB
		if err = s.tokenRepo.CreateRefresh(txCtx, refresh); err != nil {
			log.Error("failed to create refresh token", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &authV1.LoginResponse{
		AccessToken:  tokenPair.AccessToken.Value,
		RefreshToken: tokenPair.RefreshToken.Value,
		ExpiresIn:    int64(tokenPair.AccessToken.ExpiresIn.Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *authV1.LogoutRequest) (*emptypb.Empty, error) {
	log := logger.Get().SetOp("Logout")

	body := request.LogoutUser{
		RefreshToken: req.GetRefreshToken(),
	}

	//	Normalize request body
	body.Normalize()

	//	Validate request body
	if err := body.Validate(); err != nil {
		if !errs.IsValidation(err) {
			log.Error("internal validation error", zap.Error(err))

			return nil, status.Error(codes.Internal, message.ErrInternal)
		}

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	hash := s.tokenManager.HashToken(body.RefreshToken)

	//	Mark refresh token as revoked
	if err := s.tokenRepo.MarkRefreshAsRevokedByHash(ctx, hash); err != nil {
		if errors.Is(err, errs.ErrRefreshTokenNotFound) {
			//	Provided refresh token is invalid
			return &emptypb.Empty{}, status.Error(codes.Unauthenticated, message.ErrRefreshTokenInvalid)
		}

		log.Error("failed to mark refresh token as revoked", zap.Error(err))

		return &emptypb.Empty{}, status.Error(codes.Internal, message.ErrInternal)
	}

	return &emptypb.Empty{}, nil
}

func (s *AuthService) Refresh(
	ctx context.Context,
	req *authV1.RefreshRequest,
) (*authV1.RefreshResponse, error) {
	log := logger.Get().SetOp("Refresh")

	body := request.LogoutUser{
		RefreshToken: req.GetRefreshToken(),
	}

	//	Normalize request body
	body.Normalize()

	//	Validate request body
	if err := body.Validate(); err != nil {
		if !errs.IsValidation(err) {
			log.Error("internal validation error", zap.Error(err))

			return nil, status.Error(codes.Internal, message.ErrInternal)
		}

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var tokenPair jwt.TokenPair

	err := s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := ctxlib.NewTXCtx(ctx, tx)

		hash := s.tokenManager.HashToken(body.RefreshToken)

		//	Mark old refresh token as revoked
		if err := s.tokenRepo.MarkRefreshAsRevokedByHash(txCtx, hash); err != nil {
			if errors.Is(err, errs.ErrRefreshTokenNotFound) {
				return status.Error(codes.Unauthenticated, message.ErrRefreshTokenInvalid)
			}

			log.Error("failed to mark refresh token as revoked", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		}

		//	Get old refresh from DB with associated user data
		dbRefresh, err := s.tokenRepo.GetRefreshByHashWithUser(txCtx, hash)
		if err != nil {
			if errors.Is(err, errs.ErrRefreshTokenNotFound) {
				return status.Error(codes.Unauthenticated, message.ErrRefreshTokenInvalid)
			}

			log.Error("failed to find refresh by hash", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		}

		claims := jwt.Claims{
			"user_id": dbRefresh.User.ID,
			"name":    dbRefresh.User.Name,
			"email":   dbRefresh.User.Email,
		}

		//	Generate JWT token pair
		tokenPair, err = s.tokenManager.Generate(claims)
		if err != nil {
			log.Error("failed to generate jwt token pair", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		}

		refresh := &models.RefreshToken{
			UserID:    dbRefresh.User.ID,
			TokenHash: tokenPair.RefreshToken.Hash,
			ExpiresAt: dbRefresh.ExpiresAt,
		}

		//	Insert new refresh token into DB
		if err = s.tokenRepo.CreateRefresh(txCtx, refresh); err != nil {
			log.Error("failed to create refresh token", zap.Error(err))

			return status.Error(codes.Internal, message.ErrInternal)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &authV1.RefreshResponse{
		AccessToken:  tokenPair.AccessToken.Value,
		RefreshToken: tokenPair.RefreshToken.Value,
	}, nil
}
