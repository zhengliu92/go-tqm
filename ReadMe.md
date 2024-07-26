# Task Queue Manager

## Overview

The Task Queue Manager is a Go library for managing and executing concurrent tasks with configurable concurrency levels. It allows you to define multiple task queues, each with its own concurrency limit, and manage tasks within these queues. The manager provides features for adding, starting, stopping, and monitoring tasks.

## Features

- Create and manage multiple task queues.
- Define tasks with custom actions.
- Configure concurrency levels for each queue.
- Start and stop tasks with timeout control.
- Monitor task statuses and get detailed task information.
- Ensure thread-safe operations with synchronization mechanisms.

## Usage

### 1. Creating a Task

To create a task, define an action function and use the NewTask function:

```go
action := func() error {
    // Your task logic here
    return nil
}
task := NewTask("myTask", action)
```

### 2. Adding a Task to a Queue

To add a task to a queue, get the queue by name and use the AddTask method:

```go
queue, err := ITaskQueueManager.GetQueueByName("default")
if err != nil {
    // Handle error
}
err = queue.AddTask(task)
if err != nil {
    // Handle error
}
```

### 3.Starting a Task

To start a task by name, use the StartTaskByName method of the queue:

```go
err = queue.StartTaskByName("myTask", 10*time.Second)
if err != nil {
    // Handle error
}
```

### 4. Stopping a Task

To stop a running task, use the StopTaskByName method of the TaskQueueManager:

```go
err = ITaskQueueManager.StopTaskByName("myTask")
if err != nil {
    // Handle error
}
```

### 5. Getting Task Information

To get information about all tasks in a queue, use the GetTasksInfo method:

```go
tasksInfo := queue.GetTasksInfo()
for _, info := range tasksInfo {
    fmt.Printf("Task: %s, Status: %s\n", info.Name, info.Status)
}
```

To get information about all tasks across all queues, use the GetQueuesInfo method:

```go
tasksInfo := ITaskQueueManager.GetQueuesInfo()
for _, info := range *tasksInfo {
    fmt.Printf("Queue: %s, Task: %s, Status: %s, Error: %s\n", info.Queue, info.Name, info.Status, info.ErrorMsg)
}
```

### 6.Managin#g Queues

You can add new queues with specific concurrency levels:

```go
newQueue := NewQueue("custom", 15)
err := ITaskQueueManager.AddQueue(newQueue)
if err != nil {
    // Handle error
}
```

You can also list all queue names:

```go
queueNames := ITaskQueueManager.ListQueueNames()
for _, name := range queueNames {
    fmt.Println(name)
}
```
