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
		UserID:      r.GetUserId(),
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
		TaskID:      r.GetTaskId(),
		UserID:      r.GetUserId(),
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

//nolint:dupl
func (s *TaskService) SoftDelete(ctx context.Context, r *taskV1.SoftDeleteRequest) (*emptypb.Empty, error) {
	log := logger.Get().SetOp("service.task.SoftDelete")

	body := &request.SoftDeleteTask{
		TaskID: r.GetTaskId(),
		UserID: r.GetUserId(),
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

	//	Soft delete task
	if err := s.taskRepo.SoftDelete(ctx, task); err != nil {
		if errors.Is(err, errs.ErrTaskNotFound) {
			return nil, status.Error(codes.NotFound, message.ErrTaskNotFound)
		}

		log.Error("failed to delete task", zap.Error(err))

		return nil, status.Error(codes.Internal, message.ErrInternal)
	}

	return nil, nil
}

func (s *TaskService) BatchDelete(ctx context.Context, r *taskV1.BatchDeleteRequest) (*emptypb.Empty, error) {
	log := logger.Get().SetOp("service.task.BatchDelete")

	body := &request.BatchDeleteTask{
		IDs: r.GetIds(),
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

	//	Batch delete tasks by ids
	if err := s.taskRepo.BatchDelete(ctx, body.IDs); err != nil {
		if errors.Is(err, errs.ErrTaskNotFound) {
			return nil, status.Error(codes.NotFound, message.ErrTaskNotFound)
		}

		log.Error("failed to delete task", zap.Error(err))

		return nil, status.Error(codes.Internal, message.ErrInternal)
	}

	return nil, nil
}

//nolint:dupl
func (s *TaskService) Restore(ctx context.Context, r *taskV1.RestoreRequest) (*emptypb.Empty, error) {
	log := logger.Get().SetOp("service.task.Restore")

	body := &request.RestoreTask{
		TaskID: r.GetTaskId(),
		UserID: r.GetUserId(),
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
	if err := s.taskRepo.Restore(ctx, task); err != nil {
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
		UserID: r.GetUserId(),
		Type:   models.TaskType(r.GetType()),
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
	list, err := s.taskRepo.List(ctx, body.UserID, body.Type)
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
			Id:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			CreatedAt:   timestamppb.New(task.CreatedAt),
			CompletedAt: completedAt,
		})
	}

	return &taskV1.ListResponse{Task: tasks}, nil
}

//nolint:dupl
func (s *TaskService) ToggleCompletion(
	ctx context.Context,
	r *taskV1.ToggleCompletionRequest,
) (*emptypb.Empty, error) {
	log := logger.Get().SetOp("service.task.ToggleCompletion")

	body := &request.ToggleCompletion{
		TaskID: r.GetTaskId(),
		UserID: r.GetUserId(),
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
