package tqm

import (
	"fmt"
	"time"
)

func (t *TaskQueueManager) GetQueueByName(queueName string) (*TaskQueue, error) {
	if queue, exists := t.Queues[queueName]; exists {
		return queue, nil
	}
	return nil, fmt.Errorf("queue %s does not exist", queueName)
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

func NewTask(name string, action Action) *Task {
	return &Task{
		Name:    name,
		Status:  Pending,
		Action:  action,
		updated: make(chan struct{}, 1),
	}
}

func (t *TaskQueueManager) DeleteTask(taskName string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, queue := range t.Queues {
		if task, exists := queue.Tasks[taskName]; exists {
			if task.Status == Running {
				return fmt.Errorf("task %s is running, stop it first", taskName)
			}
			delete(queue.Tasks, taskName)
			return nil
		}
	}
	return fmt.Errorf("task %s does not exist", taskName)
}

func (tqm *TaskQueueManager) WaitAll() {
	tqm.wg.Wait()
}

func (tq *TaskQueueManager) StopTaskByName(taskName string) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	for _, queue := range tq.Queues {
		if task, exists := queue.Tasks[taskName]; exists {
			if task.Status != Running {
				return fmt.Errorf("task %s is not running", taskName)
			}
			return task.Stop()
		}
	}
	return fmt.Errorf("task %s does not exist", taskName)
}

func (tq *TaskQueueManager) StopAllTasks() error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	for _, queue := range tq.Queues {
		for _, task := range queue.Tasks {
			if task.Status == Running {
				if err := task.Stop(); err != nil {
					return err
				}
			}
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
	queue.updated = make(chan struct{}, 1)
	tqm.Queues[queue.Name] = queue
	tqm.Total++
	go queue.IsAnyTaskUpdated() // Start the task update watcher for the queue
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
		updated:     make(chan struct{}, 1),
	}
}

func (tqm *TaskQueueManager) IsAnyTaskUpdate() {
	for range time.Tick(time.Second) {
		tqm.mu.Lock()
		for _, queue := range tqm.Queues {
			select {
			case <-queue.updated:
				fmt.Printf("Queue %s has task updates\n", queue.Name)
				tqm.n_updates = (tqm.n_updates + 1) % 10
			default:
			}
		}
		tqm.mu.Unlock()
	}
}

func NewTaskQueueManager() *TaskQueueManager {
	manager := &TaskQueueManager{
		Queues: make(map[string]*TaskQueue),
	}
	manager.AddQueue(NewQueue("default", 10))
	manager.AddQueue(NewQueue("low", 20))
	manager.AddQueue(NewQueue("high", 5))
	go manager.IsAnyTaskUpdate()
	return manager
}
