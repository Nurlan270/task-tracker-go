package service

import (
	"context"
	"testing"
	"time"

	"github.com/Nurlan270/task-tracker-go/common/config"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/common/testutils/suite"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	appcfg "github.com/Nurlan270/task-tracker-go/task-service/internal/config"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/grpc/message"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/models"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/repository"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/repository/postgres"

	pg "github.com/Nurlan270/task-tracker-go/common/testutils/database/postgres"
	taskV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/task/v1"
	gpv "github.com/go-playground/validator/v10"
	testsuite "github.com/stretchr/testify/suite"
)

type TaskServiceTestSuite struct {
	suite.CommonSuite

	validate *validator.Validate

	taskRepo repository.TaskRepository

	taskSvc *TaskService
}

func (s *TaskServiceTestSuite) SetupSuite() {
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
		&models.Task{},
	})

	s.SetupCommonSuite(ctx)

	//	Repositories
	s.taskRepo = postgres.NewTaskRepository(s.DB())

	//	Validator
	s.validate = validator.New(gpv.WithRequiredStructEnabled())

	//	Service
	s.taskSvc = NewTaskService(s.validate, s.taskRepo)
}

func (s *TaskServiceTestSuite) SetupTest() {
	s.SetupCommonTest()
}

func (s *TaskServiceTestSuite) TearDownTest() {
	s.TearDownCommonTest()
}

func (s *TaskServiceTestSuite) SetupSubTest() {
	s.SetupCommonTest()
}

func (s *TaskServiceTestSuite) TearDownSubTest() {
	s.TearDownCommonTest()
}

func TestTaskServiceTestSuite(t *testing.T) {
	testsuite.Run(t, new(TaskServiceTestSuite))
}

func (s *TaskServiceTestSuite) TestCreate() {
	s.Run("it should validate request body", func() {
		//	Tests table
		var tests = map[string]struct {
			Title, Description string
		}{
			"long title": {
				Title:       gofakeit.LetterN(105), // 105 chars
				Description: "Some description",
			},
			"long description": {
				Title:       "Some title",
				Description: gofakeit.LetterN(5005), // 5005 chars
			},
			"empty title": {
				Title:       "",
				Description: "Some description",
			},
			"empty description": {
				Title:       "Some title",
				Description: "",
			},
		}

		for name, tt := range tests {
			s.Run(name, func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				resp, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
					UserId:      1,
					Title:       tt.Title,
					Description: tt.Description,
				})

				s.Require().Error(err)
				s.Require().Nil(resp)
				s.Require().Equal(codes.InvalidArgument, status.Code(err))
			})
		}
	})
}

func (s *TaskServiceTestSuite) TestUpdate() {
	s.Run("it should not update not existing task", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	We didn't create any task yet
		resp, err := s.taskSvc.Update(ctx, &taskV1.UpdateRequest{
			TaskId:      100,
			UserId:      5,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.NotFound, status.Code(err))
		s.Require().ErrorContains(err, message.ErrTaskNotFound)
	})
}

func (s *TaskServiceTestSuite) TestList() {
	s.Run("it should not show soft deleted task in tasks list", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Create task first
		resp, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
			UserId:      1,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().NoError(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Delete task
		resp, err = s.taskSvc.Delete(ctx, &taskV1.DeleteRequest{
			TaskId:  1,
			UserId:  1,
			IsForce: false,
		})

		s.Require().NoError(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Get tasks list
		list, err := s.taskSvc.List(ctx, &taskV1.ListRequest{
			UserId: 1,
		})

		s.Require().NoError(err)
		s.Require().Nil(list)
		s.Require().Equal(codes.OK, status.Code(err))
	})
}

func (s *TaskServiceTestSuite) TestToggleCompletion() {
	s.Run("it should not set not existing task's completion time", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	We didn't create any task yet
		resp, err := s.taskSvc.ToggleCompletion(ctx, &taskV1.ToggleCompletionRequest{
			TaskId: 100,
			UserId: 5,
		})

		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.NotFound, status.Code(err))
		s.Require().ErrorContains(err, message.ErrTaskNotFound)
	})

	s.Run("it should set completion time to nil on second call", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Create task
		createReq, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
			UserId:      5,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().NoError(err)
		s.Require().Nil(createReq)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Toggle completion to current time
		firstToggle, err := s.taskSvc.ToggleCompletion(ctx, &taskV1.ToggleCompletionRequest{
			TaskId: 1,
			UserId: 5,
		})

		s.Require().NoError(err)
		s.Require().Nil(firstToggle)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Get just updated task and assert that it has no nil completed_at field
		firstList, err := s.taskRepo.List(ctx, 5)

		s.Require().NoError(err)
		s.Require().NotEmpty(firstList)
		s.Require().NotNil(firstList[0].CompletedAt)

		//	Toggle completion to nil
		secondToggle, err := s.taskSvc.ToggleCompletion(ctx, &taskV1.ToggleCompletionRequest{
			TaskId: 1,
			UserId: 5,
		})

		s.Require().NoError(err)
		s.Require().Nil(secondToggle)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Get just updated task and assert that it has nil completed_at field
		secondList, err := s.taskRepo.List(ctx, 5)

		s.Require().NoError(err)
		s.Require().NotEmpty(secondList)
		s.Require().Nil(secondList[0].CompletedAt)
	})
}

func (s *TaskServiceTestSuite) TestDelete() {
	s.Run("it should not delete not existing task", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	We didn't create any task yet
		resp, err := s.taskSvc.Delete(ctx, &taskV1.DeleteRequest{
			TaskId:  100,
			UserId:  5,
			IsForce: true,
		})

		s.Require().Error(err)
		s.Require().Nil(resp)
		s.Require().Equal(codes.NotFound, status.Code(err))
		s.Require().ErrorContains(err, message.ErrTaskNotFound)
	})
}
