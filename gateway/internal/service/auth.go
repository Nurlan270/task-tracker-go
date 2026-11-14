package service

import (
	"context"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/request"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/response"
	authV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/auth/v1"
)

type AuthService struct {
	client authV1.AuthServiceClient
}

func NewAuthService(client authV1.AuthServiceClient) *AuthService {
	return &AuthService{
		client: client,
	}
}

func (s *AuthService) Register(ctx context.Context, req *request.Register) (response.Register, error) {
	resp, err := s.client.Register(ctx, &authV1.RegisterRequest{
		Name:                 req.Name,
		Email:                req.Email,
		Password:             req.Password,
		PasswordConfirmation: req.PasswordConfirmation,
	})

	if err != nil {
		return response.Register{}, err
	}

	//todo: here we'll send email to user through email service in future

	return response.Register{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn.AsDuration(),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *request.Login) (response.Login, error) {
	resp, err := s.client.Login(ctx, &authV1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return response.Login{}, err
	}

	return response.Login{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn.AsDuration(),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *request.Logout) error {
	_, err := s.client.Logout(ctx, &authV1.LogoutRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) Refresh(ctx context.Context, req *request.Refresh) (response.Refresh, error) {
	resp, err := s.client.Refresh(ctx, &authV1.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		return response.Refresh{}, err
	}

	return response.Refresh{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn.AsDuration(),
	}, nil
}

func (s *AuthService) GetUser(ctx context.Context, req *request.GetUser) (response.GetUser, error) {
	resp, err := s.client.GetUser(ctx, &authV1.GetUserRequest{
		AccessToken: req.AccessToken,
	})

	if err != nil {
		return response.GetUser{}, err
	}

	return response.GetUser{
		ID:    uint(resp.Id),
		Name:  resp.Name,
		Email: resp.Email,
	}, nil
}
