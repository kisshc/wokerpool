# wokerpool
Worker pool is a safe groutine pool that supports timeout control

## Base use case
```go
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
    w.Shutdown()
}
```

### Timeout control

TestPipeFull When the channel is full the message will be discarded after `3s`,Process will return a `ErrProcessTimeout` error

```go
func TestPipeFull(t *testing.T) {
	w := NewWorker()
	w.HandleWork(0, 1, 2*time.Second, func(ctx context.Context, data interface{}) {
		// delay 5s
		time.Sleep(5 * time.Second)
		t.Log(ctx.Deadline())
		t.Log(data)
	})
	w.Run()

	wg := sync.WaitGroup{}

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
	w.Shutdown()
}
```

## Panic control

TestWorkerPanic The worker does not exit when it encounters a panic error. It will print a message

```go
func TestWorkerPanic(t *testing.T)  {
	w := NewWorker()
	w.HandleWork(0, 1, 2*time.Second, func(ctx context.Context, data interface{}) {
		panic("panic err")
		t.Log(ctx.Deadline())
		t.Log(data)
	})
	w.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	w.Process(ctx,"test worker panic")
	w.Shutdown()
}
```

## Safe shutdown

TestWorkerPanic Woker waits for processing to finish and closes

```go
func TestWaitWorker(t *testing.T)  {
	w := NewWorker()
	w.HandleWork(0, 1, 2*time.Second, func(ctx context.Context, data interface{}) {
		time.Sleep(5*time.Second)
		t.Log(ctx.Deadline())
		t.Log(data)
	})
	w.Run()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	w.Process(ctx,"send message")

	// shutdown worker
	w.Shutdown()

	// 5s after print
	t.Log("worker shutdown")
}
```