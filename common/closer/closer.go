package closer

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/Nurlan270/task-tracker-go/common/logger"
)

type closeFn struct {
	name string
	fn   func() error
}

type closer struct {
	mu    sync.Mutex
	once  sync.Once
	funcs []closeFn
}

// Global closer.
var globalCloser = &closer{}

// Add adds new resource into closer's resource list.
func Add(name string, fn func() error) {
	globalCloser.add(name, fn)
}

func (c *closer) add(name string, fn func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, closeFn{name: name, fn: fn})
}

// CloseAll closes all resources within closer's resource list in "LIFO" order.
func CloseAll(ctx context.Context) {
	globalCloser.closeAll(ctx)
}

func (c *closer) closeAll(ctx context.Context) {
	c.once.Do(func() {
		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			return
		}

		log := logger.Get().SetOp("Close All")

		log.Info("Started closing resources", zap.Int("count", len(funcs)))

		//	LIFO
		for i := len(funcs) - 1; i >= 0; i-- {
			// Context of current resource; with timeout of 2 seconds
			// to avoid blocking the closing process
			rCtx, rCancel := context.WithTimeout(ctx, 2*time.Second)

			f := funcs[i]
			done := make(chan error, 1)

			log.Info("Closing resource", zap.String("name", f.name))

			go func() {
				done <- f.fn()
			}()

			var err error

			select {
			case err = <-done:
			case <-rCtx.Done():
				err = rCtx.Err()
			}

			if err != nil {
				log.Error("Failed to close resource", zap.String("name", f.name), zap.Error(err))
			} else {
				log.Info("Resource closed", zap.String("name", f.name))
			}

			rCancel()
		}
	})
}
