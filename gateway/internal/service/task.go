package service

import (
	"context"
	"time"

	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/request"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/response"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/entity"

	taskV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/task/v1"
)

type TaskService struct {
	client taskV1.TaskServiceClient
}

func NewTaskService(client taskV1.TaskServiceClient) *TaskService {
	return &TaskService{client: client}
}

func (s *TaskService) Create(ctx context.Context, req *request.CreateTask) error {
	_, err := s.client.Create(ctx, &taskV1.CreateRequest{
		UserId:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) Update(ctx context.Context, req *request.UpdateTask) error {
	_, err := s.client.Update(ctx, &taskV1.UpdateRequest{
		TaskId:      req.TaskID,
		UserId:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) List(ctx context.Context, req *request.ListTask) (response.ListTask, error) {
	list, err := s.client.List(ctx, &taskV1.ListRequest{
		UserId: req.UserID,
		Type:   req.Type.ToProto(),
	})
	if err != nil {
		return response.ListTask{}, err
	}

	tasks := make([]*entity.Task, 0, len(list.Tasks))

	//	Convert from gRPC
	for _, task := range list.Tasks {
		var completedAt *time.Time

		if task.CompletedAt != nil {
			t := task.CompletedAt.AsTime()
			completedAt = &t
		}

		tasks = append(tasks, &entity.Task{
			ID:          task.Id,
			Title:       task.Title,
			Description: task.Description,
			CreatedAt:   task.CreatedAt.AsTime(),
			CompletedAt: completedAt,
		})
	}

	return response.ListTask{Tasks: tasks}, nil
}

func (s *TaskService) ToggleCompletion(ctx context.Context, req *request.ToggleCompletion) error {
	_, err := s.client.ToggleCompletion(ctx, &taskV1.ToggleCompletionRequest{
		TaskId: req.TaskID,
		UserId: req.UserID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) SoftDelete(ctx context.Context, req *request.SoftDeleteTask) error {
	_, err := s.client.SoftDelete(ctx, &taskV1.SoftDeleteRequest{
		TaskId: req.TaskID,
		UserId: req.UserID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) BatchDelete(ctx context.Context, req *request.BatchDeleteTask) error {
	_, err := s.client.BatchDelete(ctx, &taskV1.BatchDeleteRequest{
		Ids:    req.IDs,
		UserId: req.UserID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) Restore(ctx context.Context, req *request.RestoreTask) error {
	_, err := s.client.Restore(ctx, &taskV1.RestoreRequest{
		TaskId: req.TaskID,
		UserId: req.UserID,
	})
	if err != nil {
		return err
	}

	return nil
}
