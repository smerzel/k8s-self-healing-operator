package repository

import "context"

type Task struct {
	ID      int    `json:"id"`
	User    string `json:"user"`
	Product string `json:"product"`
	Amount  int    `json:"amount"`
	Status  string `json:"status"`
}

// TaskRepository מגדיר את הפעולות מול תור המשימות ב-Redis
type TaskRepository interface {
	PushTask(ctx context.Context, task Task) error
	GetPendingTasksCount(ctx context.Context) (int64, error)
	Close() error
}