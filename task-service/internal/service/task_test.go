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
	cv "github.com/Nurlan270/task-tracker-go/task-service/internal/validator"

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
	valid := validator.New(gpv.WithRequiredStructEnabled())

	valid.RegisterCustomRules(cv.TaskTypeCustomRule())
	valid.RegisterCustomMessages(cv.Messages)

	s.validate = valid

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
		}

		for name, tt := range tests {
			s.Run(name, func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				_, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
					UserId:      1,
					Title:       tt.Title,
					Description: tt.Description,
				})

				s.Require().Error(err)
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
		_, err := s.taskSvc.Update(ctx, &taskV1.UpdateRequest{
			TaskId:      100,
			UserId:      5,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().Error(err)
		s.Require().Equal(codes.NotFound, status.Code(err))
		s.Require().ErrorContains(err, message.ErrTaskNotFound)
	})
}

func (s *TaskServiceTestSuite) TestList() {
	s.Run("it should return list of deleted tasks", func() { //nolint:dupl
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Create task first
		_, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
			UserId:      1,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Delete it softly (move to trash)
		_, err = s.taskSvc.SoftDelete(ctx, &taskV1.SoftDeleteRequest{
			UserId: 1,
			TaskId: 1,
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Get the list of deleted tasks
		list, err := s.taskSvc.List(ctx, &taskV1.ListRequest{
			UserId: 1,
			Type:   taskV1.TaskType_DELETED,
		})

		s.Require().NoError(err)
		s.Require().Len(list.Tasks, 1)
		s.Require().EqualValues(1, list.Tasks[0].Id)
	})

	s.Run("it should return list of completed tasks", func() { //nolint:dupl
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Create task first
		_, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
			UserId:      1,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Complete task
		_, err = s.taskSvc.ToggleCompletion(ctx, &taskV1.ToggleCompletionRequest{
			UserId: 1,
			TaskId: 1,
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Get the list of completed tasks
		list, err := s.taskSvc.List(ctx, &taskV1.ListRequest{
			UserId: 1,
			Type:   taskV1.TaskType_COMPLETED,
		})

		s.Require().NoError(err)
		s.Require().Len(list.Tasks, 1)
		s.Require().EqualValues(1, list.Tasks[0].Id)
	})

	s.Run("it should return list of uncompleted tasks", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Create task first
		_, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
			UserId:      1,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Get the list of uncompleted tasks
		list, err := s.taskSvc.List(ctx, &taskV1.ListRequest{
			UserId: 1,
			Type:   taskV1.TaskType_UNCOMPLETED,
		})

		s.Require().NoError(err)
		s.Require().Len(list.Tasks, 1)
		s.Require().EqualValues(1, list.Tasks[0].Id)
	})

	s.Run("it should not return list of deleted tasks in (un)/completed list", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Create task first
		_, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
			UserId:      1,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Delete it softly (move to trash)
		_, err = s.taskSvc.SoftDelete(ctx, &taskV1.SoftDeleteRequest{
			UserId: 1,
			TaskId: 1,
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Get the list of uncompleted tasks
		uncompleted, err := s.taskSvc.List(ctx, &taskV1.ListRequest{
			UserId: 1,
			Type:   taskV1.TaskType_UNCOMPLETED,
		})

		s.Require().NoError(err)
		s.Require().Nil(uncompleted)

		//	Get the list of completed tasks
		completed, err := s.taskSvc.List(ctx, &taskV1.ListRequest{
			UserId: 1,
			Type:   taskV1.TaskType_COMPLETED,
		})

		s.Require().NoError(err)
		s.Require().Nil(completed)
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
		_, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
			UserId:      5,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Toggle completion to current time
		_, err = s.taskSvc.ToggleCompletion(ctx, &taskV1.ToggleCompletionRequest{
			TaskId: 1,
			UserId: 5,
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Get just updated task and assert that it has no nil completed_at field
		firstList, err := s.taskRepo.List(ctx, 5, models.TaskTypeCompleted)

		s.Require().NoError(err)
		s.Require().NotEmpty(firstList)
		s.Require().NotNil(firstList[0].CompletedAt)

		//	Toggle completion to nil
		_, err = s.taskSvc.ToggleCompletion(ctx, &taskV1.ToggleCompletionRequest{
			TaskId: 1,
			UserId: 5,
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Get just updated task and assert that it has nil completed_at field
		secondList, err := s.taskRepo.List(ctx, 5, models.TaskTypeUncompleted)

		s.Require().NoError(err)
		s.Require().NotEmpty(secondList)
		s.Require().Nil(secondList[0].CompletedAt)
	})
}

func (s *TaskServiceTestSuite) TestSoftDelete() {
	s.Run("it should not delete task from DB", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Create task
		_, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
			UserId:      5,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Soft delete it
		_, err = s.taskSvc.SoftDelete(ctx, &taskV1.SoftDeleteRequest{
			TaskId: 1,
			UserId: 5,
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		var exists models.Task

		s.DB().
			WithContext(ctx).
			Model(&models.Task{}).
			Unscoped().
			Where("id = ?", 1).
			Where("user_id = ?", 5).
			First(&exists)

		s.Require().NotEmpty(exists)
	})

	s.Run("it should return not found on not existing task", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Soft delete not existing task
		_, err := s.taskSvc.SoftDelete(ctx, &taskV1.SoftDeleteRequest{
			TaskId: 1,
			UserId: 5,
		})

		s.Require().Error(err)
		s.Require().Equal(codes.NotFound, status.Code(err))
		s.Require().ErrorContains(err, message.ErrTaskNotFound)
	})
}

func (s *TaskServiceTestSuite) TestBatchDelete() {
	s.Run("it should return not found on not existing task", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	We didn't create any task yet
		_, err := s.taskSvc.BatchDelete(ctx, &taskV1.BatchDeleteRequest{
			Ids:    []uint64{100},
			UserId: 5,
		})

		s.Require().Error(err)
		s.Require().Equal(codes.NotFound, status.Code(err))
		s.Require().ErrorContains(err, message.ErrTaskNotFound)
	})
}

func (s *TaskServiceTestSuite) TestRestore() {
	s.Run("it should remove task from trash", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	Create task
		_, err := s.taskSvc.Create(ctx, &taskV1.CreateRequest{
			UserId:      5,
			Title:       "Some title",
			Description: "Some description",
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Soft delete it
		_, err = s.taskSvc.SoftDelete(ctx, &taskV1.SoftDeleteRequest{
			TaskId: 1,
			UserId: 5,
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Restore it
		_, err = s.taskSvc.Restore(ctx, &taskV1.RestoreRequest{
			TaskId: 1,
			UserId: 5,
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))

		//	Get a list of deleted tasks
		list, err := s.taskSvc.List(ctx, &taskV1.ListRequest{
			UserId: 5,
			Type:   taskV1.TaskType_DELETED,
		})

		s.Require().NoError(err)
		s.Require().Equal(codes.OK, status.Code(err))
		s.Require().Nil(list)
	})

	s.Run("it should return not found on not existing task", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		//	We didn't create any task yet
		_, err := s.taskSvc.Restore(ctx, &taskV1.RestoreRequest{
			TaskId: 1,
			UserId: 5,
		})

		s.Require().Error(err)
		s.Require().Equal(codes.NotFound, status.Code(err))
		s.Require().ErrorContains(err, message.ErrTaskNotFound)
	})
}
