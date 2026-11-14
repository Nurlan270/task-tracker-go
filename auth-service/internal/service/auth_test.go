package service

import (
	"context"
	"testing"
	"time"

	"github.com/Nurlan270/task-tracker-go/common/config"
	"github.com/Nurlan270/task-tracker-go/common/hasher"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/common/testutils/suite"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	appcfg "github.com/Nurlan270/task-tracker-go/auth-service/internal/config"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/grpc/message"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/jwt"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/models"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/repository"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/repository/postgres"
	cr "github.com/Nurlan270/task-tracker-go/auth-service/internal/validator"

	pg "github.com/Nurlan270/task-tracker-go/common/testutils/database/postgres"
	authV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/auth/v1"
	gpv "github.com/go-playground/validator/v10"
	testsuite "github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	suite.CommonSuite

	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository

	authSvc *AuthService

	validator    *validator.Validate
	hasher       hasher.Hasher
	tokenManager *jwt.TokenManager
}

func (s *AuthServiceTestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//	Load config
	cfg := config.MustLoadYML[appcfg.Config]("../../config/test.yml")

	//	Init noop logger
	logger.InitNoop()

	//	Setup common suite
	s.SetPostgresContainerConfig(&pg.ContainerConfig{
		Name:     cfg.Database.Name,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Img:      "postgres:18-alpine",
	})

	s.SetMigrationModels([]any{
		&models.User{}, &models.RefreshToken{},
	})

	s.SetupCommonSuite(ctx)

	//	Repositories
	s.userRepo = postgres.NewUserRepository(s.DB())
	s.tokenRepo = postgres.NewTokenRepository(s.DB())

	//	Token manager
	s.tokenManager = jwt.NewTokenManager(jwt.Config{
		AccessTTL:  cfg.JWT.AccessTTL,
		RefreshTTL: cfg.JWT.RefreshTTL,
		Secret:     cfg.JWT.Secret,
	})

	//	Hasher
	s.hasher = hasher.NewArgonHasher(nil)

	//	Validator
	valid := validator.New(gpv.WithRequiredStructEnabled())

	valid.RegisterCustomRules(cr.NameCustomRule())
	valid.RegisterCustomMessages(cr.Messages)

	s.validator = valid

	//	Service
	s.authSvc = NewAuthService(s.validator, s.tokenManager, s.hasher, s.DB(), s.userRepo, s.tokenRepo)
}

func (s *AuthServiceTestSuite) SetupTest() {
	s.SetupCommonTest()
}

func (s *AuthServiceTestSuite) TearDownTest() {
	s.TearDownCommonTest()
}

func (s *AuthServiceTestSuite) SetupSubTest() {
	s.SetupCommonTest()
}

func (s *AuthServiceTestSuite) TearDownSubTest() {
	s.TearDownCommonTest()
}

func TestAuthServiceTestSuite(t *testing.T) {
	testsuite.Run(t, new(AuthServiceTestSuite))
}

func (s *AuthServiceTestSuite) TestRegister() {
	s.Run("it should validate request body", func() {
		//	Test cases
		var tests = map[string]struct {
			Name, Email, Password, PasswordConfirmation string
		}{
			"invalid email format": {
				Email:                "not-a-valid-email",
				Name:                 "john",
				Password:             "qwerty123",
				PasswordConfirmation: "qwerty123",
			},
			"empty email": {
				Email:                "",
				Name:                 "john",
				Password:             "qwerty123",
				PasswordConfirmation: "qwerty123",
			},
			"empty password": {
				Email:                "john@mail.com",
				Name:                 "john",
				Password:             "",
				PasswordConfirmation: "",
			},
			"empty name": {
				Email:                "john@mail.com",
				Name:                 "",
				Password:             "qwerty123",
				PasswordConfirmation: "qwerty123",
			},
			"long name (longer than 50 chars)": {
				Email:                "john@mail.com",
				Name:                 "AlexanderConstantinovichMontgomeryWellingtonTheThird", // 54 chars
				Password:             "qwerty123",
				PasswordConfirmation: "qwerty123",
			},
			"invalid name format": {
				Email:                "john@mail.com",
				Name:                 "some:illegal&name;",
				Password:             "qwerty123",
				PasswordConfirmation: "qwerty123",
			},
			"too short password (less than 8 chars)": {
				Email:                "john@mail.com",
				Name:                 "john",
				Password:             "qwerty",
				PasswordConfirmation: "qwerty",
			},
			"not identical password confirmation": {
				Email:                "john@mail.com",
				Name:                 "john",
				Password:             "qwerty123",
				PasswordConfirmation: "not-identical-password",
			},
		}

		//	Run
		for name, tt := range tests {
			s.Run(name, func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				resp, err := s.authSvc.Register(ctx, &authV1.RegisterRequest{
					Name:                 tt.Name,
					Email:                tt.Email,
					Password:             tt.Password,
					PasswordConfirmation: tt.PasswordConfirmation,
				})

				s.Require().Error(err)
				s.Require().Nil(resp)
				s.Require().Equal(codes.InvalidArgument, status.Code(err))
			})
		}
	})

	s.Run("it should not register existing user", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	User
		var user = &models.User{
			Name:     "John",
			Email:    "john@mail.com",
			Password: "qwerty123",
		}

		//	Insert user into DB
		_, err := s.userRepo.Create(ctx, user)

		s.Require().NoError(err)

		//	Try to register
		resp, err := s.authSvc.Register(ctx, &authV1.RegisterRequest{
			Name:                 user.Name,
			Email:                user.Email,
			Password:             user.Password,
			PasswordConfirmation: user.Password,
		})

		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.AlreadyExists, status.Code(err))
		s.Require().ErrorContains(err, message.ErrEmailTaken)
	})
}

func (s *AuthServiceTestSuite) TestLogin() {
	s.Run("it should validate request body", func() {
		//	Test cases
		var tests = map[string]struct {
			Email, Password string
		}{
			"invalid email format": {
				Email:    "not-a-valid-email",
				Password: "qwerty123",
			},
			"empty email": {
				Email:    "",
				Password: "qwerty123",
			},
			"empty password": {
				Email:    "john@mail.com",
				Password: "",
			},
			"too short password (less than 8 chars)": {
				Email:    "john@mail.com",
				Password: "qwerty",
			},
		}

		//	Run
		for name, tt := range tests {
			s.Run(name, func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				resp, err := s.authSvc.Login(ctx, &authV1.LoginRequest{
					Email:    tt.Email,
					Password: tt.Password,
				})

				s.Require().Error(err)
				s.Require().Nil(resp)
				s.Require().Equal(codes.InvalidArgument, status.Code(err))
			})
		}
	})

	s.Run("it should not login non-existent user", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		nonExistentUser := &models.User{
			Name:     "I do not exist",
			Email:    "john@doe.com",
			Password: "qwerty123",
		}

		resp, err := s.authSvc.Login(ctx, &authV1.LoginRequest{
			Email:    nonExistentUser.Email,
			Password: nonExistentUser.Password,
		})

		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.Unauthenticated, status.Code(err))
		s.Require().ErrorContains(err, message.ErrUserNotFound)
	})

	s.Run("it should not login user with invalid password", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	User
		user := &models.User{
			Name:     "James Bond",
			Email:    "james@bond.com",
			Password: "qwerty123",
		}

		var err error

		// Hash password
		user.Password, err = s.hasher.Hash(user.Password)

		s.Require().NoError(err)

		//	Insert user into DB
		_, err = s.userRepo.Create(ctx, user)

		s.Require().NoError(err)

		resp, err := s.authSvc.Login(ctx, &authV1.LoginRequest{
			Email:    user.Email,
			Password: "invalid-password",
		})

		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.Unauthenticated, status.Code(err))
		s.Require().ErrorContains(err, message.ErrUserNotFound)
	})
}

func (s *AuthServiceTestSuite) TestLogout() {
	s.Run("it should not logout with invalid refresh token", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		resp, err := s.authSvc.Logout(ctx, &authV1.LogoutRequest{
			RefreshToken: "invalid-refresh-token",
		})

		s.Require().Error(err)
		s.Require().Empty(resp)
		s.Require().Equal(codes.Unauthenticated, status.Code(err))
		s.Require().ErrorContains(err, message.ErrTokenInvalid)
	})
}

func (s *AuthServiceTestSuite) TestRefresh() {
	s.Run("it should refresh with valid refresh token", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Register so there's actually a valid refresh token in DB
		tokenPair, err := s.authSvc.Register(ctx, &authV1.RegisterRequest{
			Name:                 "John",
			Email:                "john@mail.com",
			Password:             "qwerty123",
			PasswordConfirmation: "qwerty123",
		})

		s.Require().NoError(err)
		s.Require().NotNil(tokenPair)

		expected := &authV1.RefreshResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
		}

		//	Refresh
		actual, err := s.authSvc.Refresh(ctx, &authV1.RefreshRequest{
			RefreshToken: tokenPair.RefreshToken,
		})

		s.Require().NoError(err)
		s.Require().NotNil(actual)
		s.Require().False(cmp.Equal(expected, actual, protocmp.Transform()))
	})

	s.Run("it should not refresh with invalid refresh token", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		resp, err := s.authSvc.Refresh(ctx, &authV1.RefreshRequest{
			RefreshToken: "invalid-refresh-token",
		})

		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.Unauthenticated, status.Code(err))
		s.Require().ErrorContains(err, message.ErrTokenInvalid)
	})

	s.Run("it should not refresh with revoked refresh token", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Register so there's actually a valid refresh token in DB
		tokenPair, err := s.authSvc.Register(ctx, &authV1.RegisterRequest{
			Name:                 "John",
			Email:                "john@mail.com",
			Password:             "qwerty123",
			PasswordConfirmation: "qwerty123",
		})

		s.Require().NoError(err)
		s.Require().NotNil(tokenPair)

		//	Revoke it by calling refresh
		_, err = s.authSvc.Refresh(ctx, &authV1.RefreshRequest{
			RefreshToken: tokenPair.RefreshToken,
		})

		s.Require().NoError(err)

		//	Refresh again with the same refresh token
		resp, err := s.authSvc.Refresh(ctx, &authV1.RefreshRequest{
			RefreshToken: tokenPair.RefreshToken,
		})

		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().ErrorContains(err, message.ErrTokenInvalid)
	})
}

func (s *AuthServiceTestSuite) TestGetUser() {
	s.Run("it should return user with valid access token", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	User
		user := &models.User{
			Name:     "John",
			Email:    "john@mail.com",
			Password: "qwerty123",
		}

		//	Register new user
		tokenPair, err := s.authSvc.Register(ctx, &authV1.RegisterRequest{
			Name:                 user.Name,
			Email:                user.Email,
			Password:             user.Password,
			PasswordConfirmation: user.Password,
		})

		s.Require().NoError(err)
		s.Require().NotNil(tokenPair)

		expected := &authV1.GetUserResponse{
			Name:  user.Name,
			Email: user.Email,
		}

		//	Get user
		actual, err := s.authSvc.GetUser(ctx, &authV1.GetUserRequest{
			AccessToken: tokenPair.AccessToken,
		})

		s.Require().NoError(err)
		s.Require().NotNil(actual)

		if diff := cmp.Diff(
			expected,
			actual,
			protocmp.Transform(),
			protocmp.IgnoreFields(&authV1.GetUserResponse{}, "id"),
		); diff != "" {
			s.T().Errorf("authSvc.User() mismatch (-want +got):\n%s", diff)
		}
	})

	s.Run("it should not return user with invalid access token", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	User
		user := &models.User{
			Name:     "John",
			Email:    "john@mail.com",
			Password: "qwerty123",
		}

		//	Register new user
		tokenPair, err := s.authSvc.Register(ctx, &authV1.RegisterRequest{
			Name:                 user.Name,
			Email:                user.Email,
			Password:             user.Password,
			PasswordConfirmation: user.Password,
		})

		s.Require().NoError(err)
		s.Require().NotNil(tokenPair)

		//	Get user
		resp, err := s.authSvc.GetUser(ctx, &authV1.GetUserRequest{
			AccessToken: "invalid-access-token",
		})

		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.Unauthenticated, status.Code(err))
		s.Require().ErrorContains(err, message.ErrTokenInvalid)
	})
}
