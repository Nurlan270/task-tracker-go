package handlers

import (
	"errors"
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

func (h *TaskHandler) MarkTaskAsCompleted(c *gin.Context) {
	body := validateRequest[request.MarkTaskAsCompleted](c, h.validator,
		func(r *request.MarkTaskAsCompleted) error {
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

	// fixme: this is not good because user can mark it back as uncompleted, we need to make this method
	// dynamic

	//	Mark task as completed
	if err := h.taskSvc.MarkAsCompleted(c.Request.Context(), body); err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	body := validateRequest[request.DeleteTask](c, h.validator,
		func(r *request.DeleteTask) error {
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
	if err := h.taskSvc.Delete(c.Request.Context(), body); err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// validateRequest binds passed request, and fills it with data from request.
// Then it runs the configure function with request passed into it. After that,
// it will validate all data within request and return validated request if validation
// is successful, otherwise it will return nil. It will also return nil if configure returned error.
func validateRequest[T any](c *gin.Context, v *validator.Validate, configure func(*T) error) *T {
	var req T

	//	Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		httplib.AbortWithErr(c, errs.ErrBadRequest)
		return nil
	}

	//	Set custom fields if needed before validation
	if configure != nil {
		if err := configure(&req); err != nil {
			httplib.AbortWithErr(c, err)
			return nil
		}
	}

	//	Validate request body
	if err := v.Struct(&req); err != nil {
		var validationErr *validator.ErrValidation

		if errors.As(err, &validationErr) {
			//fixme: handle ErrValidation in error handler middleware without specifying ErrBadRequest here
			httplib.AbortWithErr(c, errors.Join(errs.ErrBadRequest, validationErr))
			return nil
		}

		httplib.AbortWithErr(c, errors.Join(errs.ErrInternalServer, err))

		return nil
	}

	return &req
}
