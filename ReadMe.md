# Super Light weight Golang Task Queue Management System

This Go program defines a concurrent task management system that allows tasks to be queued, started, stopped, and managed across multiple queues.

## Overview

The program consists of several key components:

- `Task`: Represents a single task with an associated action, status, and control mechanisms.
- `TaskQueue`: Manages a collection of tasks, providing functionality to add, remove, start, and stop tasks within the queue.
- `TaskQueueManager`: Oversees multiple task queues, handling their creation and removal.

## Features

- **Concurrent Task Execution**: Tasks within a queue can run concurrently, controlled by a semaphore that limits the number of simultaneously running tasks.
- **Task Lifecycle Management**: Tasks can be individually started and stopped.
- **Queue Management**: Task queues can be dynamically added to and removed from the system, with tasks being managed within specific queues.

## Usage

### Task Structure

Each task is defined with:

- A unique ID and name.
- An action to execute.
- A status to monitor its lifecycle (Pending, Running, Stopped, Finished).

### Adding Tasks to a Queue

Tasks must be added to a task queue before they can be started. Each task queue supports a specified level of concurrency, defining how many tasks can run in parallel.

### Starting and Stopping Tasks

Tasks can be started or stopped individually by name or all at once within their respective queue. This is managed through context cancellation to handle task stoppage gracefully.

### Managing Task Queues

Task queues can be added to the system dynamically, each with its own concurrency limit. Queues manage their tasks independently and are themselves managed by a `TaskQueueManager`.

## Example

The provided `example` function demonstrates setting up a task queue manager, adding a queue, populating it with tasks, and starting and stopping these tasks.

## Building and Running

To build and run this program:

1. Ensure you have Go installed on your system.
2. Save the code to a file named `main.go`.
3. Run the command `go build main.go` to compile the program.
4. Execute `./main` to run the compiled program.

## Dependencies

This program uses the following Go packages:

- `context`: For managing task cancellation.
- `sync`: For mutexes and wait groups to handle concurrency.
- `time`: For simulating task durations.

This task management system can be adapted for various use cases requiring asynchronous task execution and management within a Go application.
