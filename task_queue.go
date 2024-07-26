package tqm

import (
	"fmt"
	"time"
)

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

func (tq *TaskQueue) AddTask(task *Task) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if itask, exists := tq.Tasks[task.Name]; exists {
		if itask.Status == Running {
			return fmt.Errorf("task %s already exists in queue %s", task.Name, tq.Name)
		} else {
			delete(tq.Tasks, task.Name)
		}
	}
	task.wg = tq.wg
	tq.Tasks[task.Name] = task
	fmt.Printf("Task %s is added to queue %s\n", task.Name, tq.Name)
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

func (tq *TaskQueue) StartTaskByName(taskName string, timeOut time.Duration) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	task, exists := tq.Tasks[taskName]
	if !exists {
		return fmt.Errorf("task %s does not exist", taskName)
	}
	return task.Start(timeOut, tq.sem)
}

func (tq *TaskQueue) StartAllTasks(timeOut time.Duration) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	for _, task := range tq.Tasks {
		if task.Status == Running {
			continue
		}
		if err := task.Start(timeOut, tq.sem); err != nil {
			return err
		}
	}
	return nil
}

func (tq *TaskQueue) IsAnyTaskUpdated() {
	for range time.Tick(time.Second) {
		tq.mu.Lock()
		for _, task := range tq.Tasks {
			select {
			case <-task.updated:
				fmt.Printf("Task %s in queue %s has been updated\n", task.Name, tq.Name)
				tq.updated <- struct{}{} // Notify queue update
			default:
			}
		}
		tq.mu.Unlock()
	}
}
