package service

import (
	"context"
	"testing"
	"time"

	"github.com/Nurlan270/task-tracker-go/common/hasher"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/common/testutils/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Nurlan270/task-tracker-go/auth-service/internal/config"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/grpc/message"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/jwt"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/models"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/repository"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/repository/postgres"

	pg "github.com/Nurlan270/task-tracker-go/common/testutils/database/postgres"
	authV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/auth/v1"
	testsuite "github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	suite.CommonSuite

	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository

	authSvc *AuthService

	hasher       hasher.Hasher
	tokenManager *jwt.TokenManager
}

func (s *AuthServiceTestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.SetPostgresContainerConfig(&pg.ContainerConfig{
		Name:     "auth",
		User:     "auth_user",
		Password: "qwerty123",
		Img:      "postgres:18-alpine",
	})

	s.SetMigrationModels([]any{
		&models.User{}, &models.RefreshToken{},
	})

	s.SetupCommonSuite(ctx)

	//	Repositories
	s.userRepo = postgres.NewUserRepository(s.DB())
	s.tokenRepo = postgres.NewTokenRepository(s.DB())

	s.tokenManager = jwt.NewTokenManager(jwt.Config{
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour, // 7 days
	})

	s.hasher = hasher.NewArgonHasher()

	//	Service
	s.authSvc = NewAuthService(s.tokenManager, s.hasher, s.DB(), s.userRepo, s.tokenRepo)

	// fixme: this is just a stub, it should be fixed
	logger.Init(config.Config{})
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
			Name, Email, Password string
		}{
			"invalid email format": {
				Email:    "not-a-valid-email",
				Name:     "john",
				Password: "qwerty123",
			},
			"empty email": {
				Email:    "",
				Name:     "john",
				Password: "qwerty123",
			},
			"empty password": {
				Email:    "john@mail.com",
				Name:     "john",
				Password: "",
			},
			"empty name": {
				Email:    "john@mail.com",
				Name:     "",
				Password: "qwerty123",
			},
			"long name (longer than 50 chars)": {
				Email:    "john@mail.com",
				Name:     "AlexanderConstantinovichMontgomeryWellingtonTheThird", // 54 chars
				Password: "qwerty123",
			},
			"invalid name format": {
				Email:    "john@mail.com",
				Name:     "some:illegal&name;",
				Password: "qwerty123",
			},
			"too short password (less than 8 chars)": {
				Email:    "john@mail.com",
				Name:     "john",
				Password: "qwerty",
			},
		}

		//	Run
		for name, tt := range tests {
			s.Run(name, func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				resp, err := s.authSvc.Register(ctx, &authV1.RegisterRequest{
					Name:     tt.Name,
					Email:    tt.Email,
					Password: tt.Password,
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
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
		})

		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.AlreadyExists, status.Code(err))
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
		s.Require().Equal(codes.NotFound, status.Code(err))
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
