package tqm

import (
	"context"
	"sync"
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
	updated  chan struct{}
}

type TaskQueue struct {
	Name        string
	Concurrency int
	Tasks       map[string]*Task
	mu          sync.Mutex
	wg          *sync.WaitGroup
	sem         chan struct{}
	updated     chan struct{}
}

type TaskQueueManager struct {
	Queues    map[string]*TaskQueue
	mu        sync.Mutex
	Total     int
	wg        sync.WaitGroup
	n_updates int
}

type TaskInfoResp struct {
	Queue    string
	Name     string
	Status   TaskStatus
	ErrorMsg string
}

type Action func() error
