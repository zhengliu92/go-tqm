package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type TaskStatus string

const (
	Pending  TaskStatus = "pending"
	Running  TaskStatus = "running"
	Stopped  TaskStatus = "stopped"
	Finished TaskStatus = "finished"
)

type Task struct {
	ID     int
	Name   string
	Status TaskStatus
	Action func()
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
	wg     *sync.WaitGroup
}

func (t *Task) Start(sem chan struct{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status == Running {
		return errors.New("task is already running")
	}
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.Status = Running
	if t.wg != nil {
		t.wg.Add(1)
	}
	go func() {
		defer func() {
			t.mu.Lock()
			defer t.mu.Unlock()
			if t.Status != Stopped {
				t.Status = Finished
			}
			if t.wg != nil {
				t.wg.Done()
			}
		}()
		sem <- struct{}{}
		t.runWithCtx(t.ctx)
		<-sem
	}()
	return nil
}

func (t *Task) runWithCtx(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		fmt.Printf("Task %s is started\n", t.Name)
		t.Action()
		close(done)
	}()
	select {
	case <-ctx.Done():
		fmt.Printf("Task %s is stopped\n", t.Name)
	case <-done:
		fmt.Printf("Task %s is completed\n", t.Name)
	}
}

func (t *Task) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status != Running {
		return errors.New("task is not running")
	}
	t.cancel()
	t.Status = Stopped
	return nil
}

type TaskQueue struct {
	Name        string           // 任务队列名称
	Concurrency int              // 并发数
	Tasks       map[string]*Task // 任务队列
	mu          sync.Mutex
	wg          *sync.WaitGroup
	sem         chan struct{}
}

func (tq *TaskQueue) AddTask(task *Task) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if _, exists := tq.Tasks[task.Name]; exists {
		return errors.New("task already exists in queue")
	}
	task.wg = tq.wg
	tq.Tasks[task.Name] = task
	return nil
}

func (tq *TaskQueue) RemoveTask(task *Task) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if _, exists := tq.Tasks[task.Name]; !exists {
		return errors.New("task does not exist in queue")
	}
	delete(tq.Tasks, task.Name)
	return nil
}

func (tq *TaskQueue) StartTaskByName(taskName string) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	task, exists := tq.Tasks[taskName]
	if !exists {
		return errors.New("task does not exist")
	}
	return task.Start(tq.sem)
}

func (tq *TaskQueue) StopTaskByName(taskName string) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	task, exists := tq.Tasks[taskName]
	if !exists {
		return errors.New("task does not exist")
	}
	return task.Stop()
}

func (tq *TaskQueue) StartAllTasks() error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	for _, task := range tq.Tasks {
		if err := task.Start(tq.sem); err != nil {
			return err
		}
	}
	return nil
}

func (tq *TaskQueue) StopAllTasks() error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	for _, task := range tq.Tasks {
		if err := task.Stop(); err != nil {
			return err
		}
	}
	return nil
}

type TaskQueueManager struct {
	Queues map[string]*TaskQueue
	mu     sync.Mutex
	Total  int
	wg     sync.WaitGroup
}

func (tqm *TaskQueueManager) AddQueue(queue *TaskQueue) error {
	tqm.mu.Lock()
	defer tqm.mu.Unlock()
	if _, exists := tqm.Queues[queue.Name]; exists {
		return errors.New("queue already exists")
	}
	queue.wg = &tqm.wg
	queue.sem = make(chan struct{}, queue.Concurrency) // Initialize the semaphore with the concurrency limit
	tqm.Queues[queue.Name] = queue
	tqm.Total++
	return nil
}

func (tqm *TaskQueueManager) RemoveQueue(queue *TaskQueue) error {
	tqm.mu.Lock()
	defer tqm.mu.Unlock()
	if _, exists := tqm.Queues[queue.Name]; !exists {
		return errors.New("queue does not exist")
	}
	delete(tqm.Queues, queue.Name)
	tqm.Total--
	return nil
}

func run_example() {
	manager := &TaskQueueManager{
		Queues: make(map[string]*TaskQueue),
	}

	queue := &TaskQueue{
		Name:        "queue1",
		Tasks:       make(map[string]*Task),
		Concurrency: 2, // Limit the number of concurrent tasks to 2

	}

	task := &Task{
		ID:   1,
		Name: "complexTask-1",
		Action: func() {
			time.Sleep(10 * time.Second)
		},
		Status: Pending,
	}

	task2 := &Task{
		ID:   2,
		Name: "complexTask-2",
		Action: func() {
			time.Sleep(10 * time.Second)
		},
		Status: Pending,
	}

	// Add queue to manager
	if err := manager.AddQueue(queue); err != nil {
		fmt.Println(err)
		return
	}

	// Add task to queue
	if err := queue.AddTask(task); err != nil {
		fmt.Println(err)
		return
	}
	// Add task to queue
	if err := queue.AddTask(task2); err != nil {
		fmt.Println(err)
		return
	}

	if err := queue.StartAllTasks(); err != nil {
		fmt.Println(err)
		return
	}

	time.Sleep(2 * time.Second)

	// Stop task by name
	if err := queue.StopTaskByName("complexTask-1"); err != nil {
		fmt.Println(err)
		return
	}
	time.Sleep(2 * time.Second)

	// Restart task by name
	if err := queue.StartTaskByName("complexTask-1"); err != nil {
		fmt.Println(err)
		return
	}

	manager.wg.Wait() // Wait for all tasks to complete}
}

func main() {
	run_example()
}
