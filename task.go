package tqm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func (t *Task) Start(timeOut time.Duration, sem chan struct{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status == Running {
		return fmt.Errorf("task %s is already running", t.Name)
	}
	// Initialize context and cancel function
	t.ctx, t.cancel = context.WithTimeout(context.Background(), timeOut)
	t.Status = Running
	t.updated <- struct{}{} // Notify task update
	if t.wg == nil {
		t.wg = &sync.WaitGroup{}
	}
	t.wg.Add(1)
	go func() {
		defer func() {
			t.mu.Lock()
			defer t.mu.Unlock()
			t.wg.Done()
		}()
		sem <- struct{}{}
		t.runWithCtx(t.ctx)
		<-sem
	}()
	return nil
}

func (t *Task) runWithCtx(ctx context.Context) {
	done := make(chan struct{})
	runErr := make(chan error)
	go func() {
		defer close(runErr)
		fmt.Printf("Task %s is started\n", t.Name)
		err := t.Action()
		if err != nil {
			runErr <- err
		} else {
			close(done)
		}
	}()
	select {
	case <-ctx.Done():
		fmt.Printf("Task %s is stopped\n", t.Name)

		t.Status = Stopped
	case <-done:
		fmt.Printf("Task %s is completed\n", t.Name)
		t.Status = Finished
		t.Action = nil
	case err := <-runErr:
		fmt.Printf("Task %s is failed: %v\n", t.Name, err)
		t.ErrorMsg = err.Error()
		t.Status = Failed
	}
	t.updated <- struct{}{} // Notify task update
}

func (t *Task) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status != Running {
		return fmt.Errorf("task %s is not running", t.Name)
	}
	t.cancel()
	t.updated <- struct{}{} // Notify task update
	return nil
}
