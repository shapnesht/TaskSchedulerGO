package main

import (
	"fmt"
	"sync"
	"time"
)

type Task struct {
	ID          string
	Name        string
	Description string
	IsRecurring bool
	stopChannel chan struct{}
	wg          sync.WaitGroup
}

type TaskScheduler struct {
	tasks    map[string]*Task
	mu       sync.Mutex
	globalWg sync.WaitGroup
}

func NewScheduler() *TaskScheduler {
	return &TaskScheduler{
		tasks: make(map[string]*Task),
	}
}

func generateID() string {
	return fmt.Sprintf("Task-%d", time.Now().Nanosecond())
}

func (t *TaskScheduler) ScheduleOnce(taskName string, taskDescription string, delay time.Duration, action func()) string {
	taskId := generateID()

	fmt.Printf("[%s]Scheduling One off task [%s] to run after %s second(s)\n", time.Now().Format("15:04:05.000"), taskId, delay)

	t.globalWg.Add(1)

	time.AfterFunc(delay, func() {
		defer t.globalWg.Done()
		fmt.Printf("[%s]Executing One off task %s\n", time.Now().Format("15:04:05.000"), taskId)
		action()
		fmt.Printf("[%s]Completed One off task %s\n", time.Now().Format("15:04:05.000"), taskId)
	})

	return taskId
}

func (t *TaskScheduler) ScheduleRecurring(taskName string, taskDescription string, interval time.Duration, action func()) (string, error) {
	taskId := generateID()

	task := &Task{
		ID:          taskId,
		Name:        taskName,
		Description: taskDescription,
		IsRecurring: true,
		stopChannel: make(chan struct{}),
	}

	t.mu.Lock()
	t.tasks[taskId] = task
	t.mu.Unlock()

	fmt.Printf("[%s]Scheduling Recurring task %s to run after every %s second(s)\n", time.Now().Format("15:04:05.000"), taskId, interval)

	t.globalWg.Add(1)
	task.wg.Add(1)

	go func() {
		defer t.globalWg.Done()
		defer task.wg.Done()

		fmt.Printf("[%s]New Goroutine started for Task %s. It'll run after initial delay of %s seconds.\n", time.Now().Format("15:04:05.000"), taskId, interval)
		initTimer := time.NewTimer(interval)

		select {
		case <-initTimer.C:

		case <-task.stopChannel:
			initTimer.Stop()
			fmt.Printf("[%s]Task %s stopped before initial run.\n", time.Now().Format("15:04:05.000"), taskId)
			return
		}

		fmt.Printf("[%s]Initial run of the task %s.\n", time.Now().Format("15:04:05.000"), taskId)
		action()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Printf("[%s]Running the recurring task %s.\n", time.Now().Format("15:04:05.000"), taskId)
				action()
			case <-task.stopChannel:
				fmt.Printf("[%s]Task %s stopped after getting a message in Stop Channel. Exiting Goroutine.\n", time.Now().Format("15:04:05.000"), taskId)
				return
			}
		}
	}()

	return taskId, nil
}

func (t *TaskScheduler) StopTask(taskId string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if task, exists := t.tasks[taskId]; exists {
		fmt.Printf("[%s]Stopping task %s\n", time.Now().Format("15:04:05.000"), taskId)
		close(task.stopChannel)
		delete(t.tasks, taskId)
		task.wg.Wait()
		fmt.Printf("[%s]Task %s stopped successfully.\n", time.Now().Format("15:04:05.000"), taskId)
		return nil
	}
	return fmt.Errorf("task not found: %s", taskId)
}

func (t *TaskScheduler) StopAllTasks() {
	t.mu.Lock()

	tasksToStop := make([]string, 0, len(t.tasks))
	for taskId := range t.tasks {
		tasksToStop = append(tasksToStop, taskId)
	}
	t.mu.Unlock()

	for _, taskId := range tasksToStop {
		t.StopTask(taskId)
	}
	t.globalWg.Wait()
}

func main() {
	taskScheduler := NewScheduler()
	taskScheduler.ScheduleOnce("3sec", "Print shapnesh after 3 sec", time.Second*3, func() {
		fmt.Printf("Shapnesh...\n")
	})

	taskScheduler.ScheduleRecurring("5sec", "Print shapnesh after every 5 sec", time.Second*5, func() {
		fmt.Printf("Shapnesh...\n")
	})
	taskScheduler.globalWg.Wait()
}
