package handlers

import (
	"net/http"
	"strconv"

	"github.com/Nurlan270/task-tracker-go/common/validator"
	"github.com/gin-gonic/gin"

	ctxlib "github.com/Nurlan270/task-tracker-go/gateway/internal/context"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/request"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/entity"
	errs "github.com/Nurlan270/task-tracker-go/gateway/internal/errors"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/service"
)

type TaskHandler struct {
	validator *validator.Validate
	taskSvc   *service.TaskService
}

func NewTaskHandler(validator *validator.Validate, taskSvc *service.TaskService) *TaskHandler {
	return &TaskHandler{
		validator: validator,
		taskSvc:   taskSvc,
	}
}

func (h *TaskHandler) Create(c *gin.Context) {
	body := validateRequest[request.CreateTask](c, h.validator,
		func(r *request.CreateTask) error {
			//	Get user id from ctx
			user := c.MustGet(ctxlib.UserCtxKey).(*entity.User)
			r.UserID = uint64(user.ID)

			return nil
		})
	if body == nil {
		return
	}

	//	Create task
	if err := h.taskSvc.Create(c.Request.Context(), body); err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *TaskHandler) Update(c *gin.Context) {
	body := validateRequest[request.UpdateTask](c, h.validator,
		func(r *request.UpdateTask) error {
			//	Get user id from ctx
			user := c.MustGet(ctxlib.UserCtxKey).(*entity.User)
			r.UserID = uint64(user.ID)

			//	Get task id from url param
			taskIdStr, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				return errs.ErrBadRequest
			}

			r.TaskID = taskIdStr

			return nil
		})
	if body == nil {
		return
	}

	//	Update task
	if err := h.taskSvc.Update(c.Request.Context(), body); err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *TaskHandler) List(c *gin.Context) {
	body := validateRequest[request.ListTask](c, h.validator,
		func(r *request.ListTask) error {
			//	Get user id from ctx
			user := c.MustGet(ctxlib.UserCtxKey).(*entity.User)
			r.UserID = uint64(user.ID)

			return nil
		})
	if body == nil {
		return
	}

	//	Get tasks list
	resp, err := h.taskSvc.List(c.Request.Context(), body)
	if err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) ToggleCompletion(c *gin.Context) {
	body := validateRequest[request.ToggleCompletion](c, h.validator,
		func(r *request.ToggleCompletion) error {
			//	Get user id from ctx
			user := c.MustGet(ctxlib.UserCtxKey).(*entity.User)
			r.UserID = uint64(user.ID)

			//	Get task id from url param
			taskIdStr, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				return errs.ErrBadRequest
			}

			r.TaskID = taskIdStr

			return nil
		},
	)
	if body == nil {
		return
	}

	//	Toggle task's completion time
	if err := h.taskSvc.ToggleCompletion(c.Request.Context(), body); err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *TaskHandler) SoftDelete(c *gin.Context) {
	body := validateRequest[request.SoftDeleteTask](c, h.validator,
		func(r *request.SoftDeleteTask) error {
			//	Get user id from ctx
			user := c.MustGet(ctxlib.UserCtxKey).(*entity.User)
			r.UserID = uint64(user.ID)

			//	Get task id from url param
			taskIdStr, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				return errs.ErrBadRequest
			}

			r.TaskID = taskIdStr

			return nil
		})
	if body == nil {
		return
	}

	//	Delete task
	if err := h.taskSvc.SoftDelete(c.Request.Context(), body); err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TaskHandler) BatchDelete(c *gin.Context) {
	body := validateRequest[request.BatchDeleteTask](c, h.validator, nil)
	if body == nil {
		return
	}

	//	Delete task
	if err := h.taskSvc.BatchDelete(c.Request.Context(), body); err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TaskHandler) Restore(c *gin.Context) {
	body := validateRequest[request.RestoreTask](c, h.validator,
		func(r *request.RestoreTask) error {
			//	Get user id from ctx
			user := c.MustGet(ctxlib.UserCtxKey).(*entity.User)
			r.UserID = uint64(user.ID)

			//	Get task id from url param
			taskIdStr, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				return errs.ErrBadRequest
			}

			r.TaskID = taskIdStr

			return nil
		})
	if body == nil {
		return
	}

	//	Restore task
	if err := h.taskSvc.Restore(c.Request.Context(), body); err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	c.Status(http.StatusOK)
}
