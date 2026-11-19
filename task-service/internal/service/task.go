package service

import (
	"context"
	"errors"

	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	errs "github.com/Nurlan270/task-tracker-go/task-service/internal/errors"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/grpc/dto/request"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/grpc/dto/response"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/grpc/message"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/models"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/repository"

	taskV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/task/v1"
)

type TaskService struct {
	taskV1.UnimplementedTaskServiceServer

	validate *validator.Validate
	taskRepo repository.TaskRepository
}

func NewTaskService(validate *validator.Validate, taskRepo repository.TaskRepository) *TaskService {
	return &TaskService{
		validate: validate,
		taskRepo: taskRepo,
	}
}

func (s *TaskService) Create(ctx context.Context, r *taskV1.CreateRequest) (*emptypb.Empty, error) {
	log := logger.Get().SetOp("service.task.Create")

	body := &request.CreateTask{
		UserID:      uint(r.GetUserId()),
		Title:       r.GetTitle(),
		Description: r.GetDescription(),
	}

	//	Validate request body
	if err := s.validate.Struct(body); err != nil {
		var validationErr *validator.ErrValidation

		if !errors.As(err, &validationErr) {
			log.Error("internal validation error", zap.Error(err))

			return nil, status.Error(codes.Internal, message.ErrInternal)
		}

		return nil, response.ToGRPCValidationErr(err)
	}

	task := &models.Task{
		UserID:      body.UserID,
		Title:       body.Title,
		Description: body.Description,
	}

	//	Create task
	if err := s.taskRepo.Create(ctx, task); err != nil {
		log.Error("failed to create task", zap.Error(err))

		return nil, status.Error(codes.Internal, message.ErrInternal)
	}

	return nil, nil
}

func (s *TaskService) Update(ctx context.Context, r *taskV1.UpdateRequest) (*emptypb.Empty, error) {
	log := logger.Get().SetOp("service.task.Update")

	body := &request.UpdateTask{
		TaskID:      uint(r.GetTaskId()),
		UserID:      uint(r.GetUserId()),
		Title:       r.GetTitle(),
		Description: r.GetDescription(),
	}

	//	Validate request body
	if err := s.validate.Struct(body); err != nil {
		var validationErr *validator.ErrValidation

		if !errors.As(err, &validationErr) {
			log.Error("internal validation error", zap.Error(err))

			return nil, status.Error(codes.Internal, message.ErrInternal)
		}

		return nil, response.ToGRPCValidationErr(err)
	}

	task := &models.Task{
		ID:          body.TaskID,
		UserID:      body.UserID,
		Title:       body.Title,
		Description: body.Description,
	}

	//	Update task
	if err := s.taskRepo.Update(ctx, task); err != nil {
		if errors.Is(err, errs.ErrTaskNotFound) {
			return nil, status.Error(codes.NotFound, message.ErrTaskNotFound)
		}

		log.Error("failed to update task", zap.Error(err))

		return nil, status.Error(codes.Internal, message.ErrInternal)
	}

	return nil, nil
}

func (s *TaskService) Delete(ctx context.Context, r *taskV1.DeleteRequest) (*emptypb.Empty, error) {
	log := logger.Get().SetOp("service.task.Delete")

	body := &request.DeleteTask{
		TaskID:  uint(r.GetTaskId()),
		UserID:  uint(r.GetUserId()),
		IsForce: r.GetIsForce(),
	}

	//	Validate request body
	if err := s.validate.Struct(body); err != nil {
		var validationErr *validator.ErrValidation

		if !errors.As(err, &validationErr) {
			log.Error("internal validation error", zap.Error(err))

			return nil, status.Error(codes.Internal, message.ErrInternal)
		}

		return nil, response.ToGRPCValidationErr(err)
	}

	task := &models.Task{
		ID:     body.TaskID,
		UserID: body.UserID,
	}

	//	Delete task
	if err := s.taskRepo.Delete(ctx, task, body.IsForce); err != nil {
		if errors.Is(err, errs.ErrTaskNotFound) {
			return nil, status.Error(codes.NotFound, message.ErrTaskNotFound)
		}

		log.Error("failed to delete task", zap.Error(err))

		return nil, status.Error(codes.Internal, message.ErrInternal)
	}

	return nil, nil
}

func (s *TaskService) List(ctx context.Context, r *taskV1.ListRequest) (*taskV1.ListResponse, error) {
	log := logger.Get().SetOp("service.task.List")

	body := &request.ListTask{
		UserID: uint(r.GetUserId()),
	}

	//	Validate request body
	if err := s.validate.Struct(body); err != nil {
		var validationErr *validator.ErrValidation

		if !errors.As(err, &validationErr) {
			log.Error("internal validation error", zap.Error(err))

			return nil, status.Error(codes.Internal, message.ErrInternal)
		}

		return nil, response.ToGRPCValidationErr(err)
	}

	//	Get tasks list
	list, err := s.taskRepo.List(ctx, body.UserID)
	if err != nil {
		log.Error("failed to get list of tasks", zap.Error(err))

		return nil, status.Error(codes.Internal, message.ErrInternal)
	}

	if len(list) == 0 {
		//	No tasks were found
		return nil, nil
	}

	tasks := make([]*taskV1.Task, 0, len(list))

	//	Convert list of tasks into list of gRPC tasks
	for _, task := range list {
		var completedAt *timestamppb.Timestamp

		if task.CompletedAt != nil {
			completedAt = timestamppb.New(*task.CompletedAt)
		}

		tasks = append(tasks, &taskV1.Task{
			Title:       task.Title,
			Description: task.Description,
			CompletedAt: completedAt,
		})
	}

	return &taskV1.ListResponse{Task: tasks}, nil
}

func (s *TaskService) ToggleCompletion(
	ctx context.Context,
	r *taskV1.ToggleCompletionRequest,
) (*emptypb.Empty, error) {
	log := logger.Get().SetOp("service.task.ToggleCompletion")

	body := &request.ToggleCompletion{
		TaskID: uint(r.GetTaskId()),
		UserID: uint(r.GetUserId()),
	}

	//	Validate request body
	if err := s.validate.Struct(body); err != nil {
		var validationErr *validator.ErrValidation

		if !errors.As(err, &validationErr) {
			log.Error("internal validation error", zap.Error(err))

			return nil, status.Error(codes.Internal, message.ErrInternal)
		}

		return nil, response.ToGRPCValidationErr(err)
	}

	task := &models.Task{
		ID:     body.TaskID,
		UserID: body.UserID,
	}

	//	Toggle task's completion time
	if err := s.taskRepo.ToggleCompletion(ctx, task); err != nil {
		if errors.Is(err, errs.ErrTaskNotFound) {
			return nil, status.Error(codes.NotFound, message.ErrTaskNotFound)
		}

		log.Error("failed to mark task as completed", zap.Error(err))

		return nil, status.Error(codes.Internal, message.ErrInternal)
	}

	return nil, nil
}
