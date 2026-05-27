package repository

import "context"

type Task struct {
	ID      int
	User    string
	Product string
	Amount  int
	Status  string
}

type TaskRepository interface {
	SaveTask(ctx context.Context, task Task) error
	GetTotalAmount(ctx context.Context, product string) (int, error)
	Ping(ctx context.Context) error
	GetPendingTasksCount(ctx context.Context) (int, error)
	GetNextPendingTask(ctx context.Context) (*Task, error)
	MarkTaskDone(ctx context.Context, id int) error
}
