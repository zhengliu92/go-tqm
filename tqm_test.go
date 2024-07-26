package tqm

import (
	"fmt"
	"testing"
	"time"
)

func Test_Tqm(t *testing.T) {
	tqm := NewTaskQueueManager()
	queue, err := tqm.GetQueueByName("default")
	if err != nil {
		t.Errorf("GetQueueByName failed: %v", err)
	}
	action_ok := func() error {
		time.Sleep(3 * time.Second)
		return nil
	}
	task_ok := NewTask("task_ok", action_ok)
	queue.AddTask(task_ok)

	queue.StartTaskByName("task_ok", 10*time.Second)

	action_fail := func() error {
		time.Sleep(2 * time.Second)
		return fmt.Errorf("action failed")
	}
	task_fail := NewTask("task_fail", action_fail)
	queue.AddTask(task_fail)
	queue.StartAllTasks(10 * time.Second)
	tqm.WaitAll()
}
