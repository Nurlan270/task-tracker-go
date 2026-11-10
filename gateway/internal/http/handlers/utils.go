package handlers

import (
	"context"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"go.uber.org/zap"
)

func safeHandleErr(ctx context.Context, err error, fn func(err error)) {
	if ctx != nil && ctx.Err() != nil {
		logger.Get().
			SetOp("safeHandleErr").
			Error("context has error. Response writer was terminated", zap.Error(ctx.Err()))
		return
	}

	if err != nil {
		fn(err)
	}
}
