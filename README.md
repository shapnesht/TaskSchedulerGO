# TaskSchedulerGO

A small in-memory task scheduler written in Go. It runs functions either once after a delay or repeatedly at a fixed interval, using goroutines, timers and tickers from the standard library only.

## Features

- **One-off tasks**: run a function once after a delay (`ScheduleOnce`).
- **Recurring tasks**: run a function every `interval`, with the first run after one interval (`ScheduleRecurring`).
- **Stop a task**: cancel a recurring task by ID (`StopTask`) and wait for its goroutine to exit.
- **Stop everything**: cancel all recurring tasks and wait for all work to finish (`StopAllTasks`).
- **Timestamped logging**: every scheduling, execution and stop event is printed with a millisecond timestamp.

## Requirements

- Go 1.18 or newer

## Run

```sh
go run main.go
```

The demo in `main()` schedules a one-off task (3s) and a recurring task (every 5s). The recurring task is never stopped there, so the program runs until you press `Ctrl+C`.

Example output:

```
[10:15:00.000]Scheduling One off task [Task-123456789] to run after 3s second(s)
[10:15:00.000]Scheduling Recurring task Task-987654321 to run after every 5s second(s)
[10:15:03.000]Executing One off task Task-123456789
Shapnesh...
[10:15:03.000]Completed One off task Task-123456789
[10:15:05.000]Initial run of the task Task-987654321.
Shapnesh...
```

## Usage

```go
scheduler := NewScheduler()

// Run once after 3 seconds
scheduler.ScheduleOnce("greet", "Say hello once", 3*time.Second, func() {
    fmt.Println("hello")
})

// Run every 5 seconds
id, err := scheduler.ScheduleRecurring("tick", "Say tick", 5*time.Second, func() {
    fmt.Println("tick")
})
if err != nil {
    log.Fatal(err)
}

// Later: stop one task, or all of them
scheduler.StopTask(id)
scheduler.StopAllTasks()
```

## API

| Method | Description |
| --- | --- |
| `NewScheduler() *TaskScheduler` | Creates an empty scheduler. |
| `ScheduleOnce(name, description string, delay time.Duration, action func()) string` | Runs `action` once after `delay`. Returns the task ID. |
| `ScheduleRecurring(name, description string, interval time.Duration, action func()) (string, error)` | Runs `action` every `interval`. Returns the task ID. |
| `StopTask(taskID string) error` | Stops a recurring task and waits for it to exit. Returns an error if the ID is unknown. |
| `StopAllTasks()` | Stops every recurring task and waits for all scheduled work to complete. |

## How it works

- Each recurring task runs in its own goroutine and listens on a per-task stop channel, so cancellation is immediate, even before the first run.
- A mutex guards the task map; a `sync.WaitGroup` per task and one global `WaitGroup` let callers wait for clean shutdown.
- Task IDs are generated from the current nanosecond (`Task-<n>`).

## Known limitations

- One-off tasks are not registered in the task map, so `StopTask` cannot cancel them.
- Task IDs come from `time.Now().Nanosecond()`, so collisions are possible when tasks are created in quick succession.
- If `action` panics, the goroutine (and the program) crashes.
- State is in memory only; nothing is persisted across restarts.
- `ScheduleRecurring` always returns a `nil` error today.

## Project layout

```
.
├── main.go    # scheduler implementation and demo
└── README.md
```
