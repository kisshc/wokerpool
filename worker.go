package worker

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"
)

var ErrProcessTimeout = errors.New("worker pipeline data sending timeout")

type handFun func(ctx context.Context, data interface{})

type IWorker interface {
	// Run implementation interface server
	Run() error
	// Shutdown implementation interface server
	Shutdown() error

	HandleWork(pipeSize int, workerPoolSize int, maxSec time.Duration, fun handFun)
	Process(ctx context.Context, data interface{}) error
}

type Worker struct {
	ctx    context.Context
	num    int
	maxSec time.Duration
	fun    func(ctx context.Context, data interface{})
	data   chan interface{}
	stop   chan struct{}
	group  *errgroup.Group

	runOnce sync.Once
}

func NewWorker() IWorker {
	g, ctx := errgroup.WithContext(context.Background())
	w := &Worker{
		ctx:   ctx,
		group: g,
		stop:  make(chan struct{}),
	}
	return w
}

func (w *Worker) HandleWork(pipeSize int, poolSize int, maxSec time.Duration, fun handFun) {
	w.num = poolSize
	w.fun = fun
	w.maxSec = maxSec
	w.data = make(chan interface{}, pipeSize)
}

func (w *Worker) Process(ctx context.Context, data interface{}) error {
	select {
	case <-ctx.Done():
		return ErrProcessTimeout
	case w.data <- data:
		return nil
	}
}

func (w *Worker) Run() error {
	w.runOnce.Do(func() {
		for i := 0; i < w.num; i++ {
			w.group.Go(func() error {
			cycle:
				for true {
					select {
					case data := <-w.data:
						func() {
							defer func() {
								if err := recover(); err != nil {
									fmt.Printf("Worker fatal error：%s\n", err)
								}
							}()
							ctxFun, cancel := context.WithTimeout(w.ctx, w.maxSec)
							w.fun(ctxFun, data)
							cancel()
						}()
					case <-w.stop:
						break cycle
					}
				}
				return nil
			})
		}
	})
	return nil
}

func (w *Worker) Shutdown() error {
	close(w.stop)
	err := w.group.Wait()
	if err != nil {
		return err
	}
	return nil
}
