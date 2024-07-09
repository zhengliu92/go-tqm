package main

import (
	"context"
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
	Failed   TaskStatus = "failed"
)

type Task struct {
	Name     string
	Status   TaskStatus
	Action   func() error
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	wg       *sync.WaitGroup
	ErrorMsg string
}

type TaskQueue struct {
	Name        string
	Concurrency int
	Tasks       map[string]*Task
	mu          sync.Mutex
	wg          *sync.WaitGroup
	sem         chan struct{}
}

type TaskQueueManager struct {
	Queues map[string]*TaskQueue
	mu     sync.Mutex
	Total  int
	wg     sync.WaitGroup
}

func (t *TaskQueueManager) GetQueueByName(queueName string) (*TaskQueue, error) {
	if queue, exists := t.Queues[queueName]; exists {
		return queue, nil
	}
	return nil, fmt.Errorf("queue %s does not exist", queueName)
}

type TaskInfoResp struct {
	Queue    string
	Name     string
	Status   TaskStatus
	ErrorMsg string
}

func (t *TaskQueue) GetTasksInfo() []TaskInfoResp {
	t.mu.Lock()
	defer t.mu.Unlock()
	tasksInfo := make([]TaskInfoResp, 0)
	for name, task := range t.Tasks {
		tasksInfo = append(tasksInfo, TaskInfoResp{
			Queue:  t.Name,
			Name:   name,
			Status: task.Status,
		})
	}
	return tasksInfo
}

func (t *TaskQueueManager) GetQueuesInfo() *[]TaskInfoResp {
	t.mu.Lock()
	defer t.mu.Unlock()
	tasksInfo := make([]TaskInfoResp, 0)
	for _, queue := range t.Queues {
		for name, task := range queue.Tasks {
			tasksInfo = append(tasksInfo, TaskInfoResp{
				Queue:    queue.Name,
				Name:     name,
				Status:   task.Status,
				ErrorMsg: task.ErrorMsg,
			})
		}
	}
	return &tasksInfo
}

func (t *TaskQueueManager) ListQueueNames() []string {
	queueNames := make([]string, 0)
	for name := range t.Queues {
		queueNames = append(queueNames, name)
	}
	return queueNames
}

func NewTask(name string, action func() error) *Task {
	return &Task{
		Name:   name,
		Status: Pending,
		Action: action,
	}
}

func (t *Task) Start(timeOut time.Duration, sem chan struct{}) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status == Running {
		return fmt.Errorf("task %s is already running", t.Name)
	}
	t.ctx, t.cancel = context.WithTimeout(context.Background(), timeOut)
	t.Status = Running
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
	run_err := make(chan error)
	go func() {
		Logrus.Infof("Task %s is started\n", t.Name)
		err := t.Action()
		if err != nil {
			run_err <- err
		}
		close(done)
	}()
	select {
	case <-ctx.Done():
		Logrus.Infof("Task %s is stopped\n", t.Name)
		t.Status = Stopped
	case <-done:
		Logrus.Infof("Task %s is completed\n", t.Name)
		t.Status = Finished
	case err := <-run_err:
		Logrus.Errorf("Task %s is failed: %v\n", t.Name, err)
		t.ErrorMsg = err.Error()
		t.Status = Failed
	}
}

func (t *Task) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status != Running {
		return fmt.Errorf("task %s is not running", t.Name)
	}
	t.cancel()
	return nil
}

func (tq *TaskQueue) AddTask(task *Task) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if itask, exists := tq.Tasks[task.Name]; exists {
		if itask.Status == Running {
			return fmt.Errorf("task %s already exists in queue %s", task.Name, tq.Name)
		} else {
			tq.RemoveTask(itask)
		}
	}
	task.wg = tq.wg
	tq.Tasks[task.Name] = task
	Logrus.Infof("Task %s is added to queue %s\n", task.Name, tq.Name)
	return nil
}

func (tq *TaskQueue) RemoveTask(task *Task) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if _, exists := tq.Tasks[task.Name]; !exists {
		return fmt.Errorf("task %s does not exist in queue %s", task.Name, tq.Name)
	}
	delete(tq.Tasks, task.Name)
	return nil
}
func (tq *TaskQueue) WaitAll() {
	tq.wg.Wait()
}

func (tqm *TaskQueueManager) WaitAll() {
	tqm.wg.Wait()
}

func (tq *TaskQueue) StartTaskByName(taskName string, timeOut time.Duration) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	task, exists := tq.Tasks[taskName]
	if !exists {
		return fmt.Errorf("task %s does not exist", taskName)
	}
	return task.Start(timeOut, tq.sem)
}

func (tq *TaskQueue) StopTaskByName(taskName string) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	task, exists := tq.Tasks[taskName]
	if !exists {
		return fmt.Errorf("task %s does not exist", taskName)
	}
	return task.Stop()
}

func (tq *TaskQueue) StartAllTasks(timeOut time.Duration) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	for _, task := range tq.Tasks {
		if err := task.Start(timeOut, tq.sem); err != nil {
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

func (tqm *TaskQueueManager) AddQueue(queue *TaskQueue) error {
	tqm.mu.Lock()
	defer tqm.mu.Unlock()
	if _, exists := tqm.Queues[queue.Name]; exists {
		return fmt.Errorf("queue %s already exists", queue.Name)
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
		return fmt.Errorf("queue %s does not exist", queue.Name)
	}
	delete(tqm.Queues, queue.Name)
	tqm.Total--
	return nil
}

func NewQueue(name string, concurrency int) *TaskQueue {
	return &TaskQueue{
		Name:        name,
		Concurrency: concurrency,
		Tasks:       make(map[string]*Task),
	}
}

func NewTaskQueueManager() *TaskQueueManager {
	manager := &TaskQueueManager{
		Queues: make(map[string]*TaskQueue),
	}
	manager.AddQueue(NewQueue("default", 10))
	manager.AddQueue(NewQueue("low", 20))
	manager.AddQueue(NewQueue("high", 5))
	return manager
}

var ITaskQueueManager = NewTaskQueueManager()
