package worker

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestWorker 正常测试
func TestWorker(t *testing.T) {
	w := NewWorker()
	w.HandleWork(1, 1, 2*time.Second, func(ctx context.Context, data interface{}) {
		t.Log(ctx.Deadline())
		t.Log(data)
	})
	w.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	w.Process(ctx, "hello worker pool")

	time.Sleep(1 * time.Second)
}

func TestPipeFull(t *testing.T) {
	w := NewWorker()
	w.HandleWork(0, 1, 2*time.Second, func(ctx context.Context, data interface{}) {
		// 处理时间增长
		time.Sleep(5 * time.Second)
		t.Log(ctx.Deadline())
		t.Log(data)
	})
	w.Run()

	wg := sync.WaitGroup{}

	// 并发放入两条任务，每个任务执行5s
	// 第一个顺利加入
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := w.Process(ctx, "hello worker pool 1")
		if err != nil {
			t.Log(err.Error())
		}
	}()

	// 三秒超时
	wg.Add(1)
	go func() {
		wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := w.Process(ctx, "hello worker pool 2")
		if err != nil {
			t.Log(err.Error())
		}
	}()

	wg.Wait()
}
