package middleware

import (
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/response"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/message"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func Recover(c *gin.Context, recovered any) {
	if p, ok := recovered.(string); ok {
		//	Log panic message
		logger.Get().Error("Recovered from panic", zap.String("panic", p))
	}

	c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
}
