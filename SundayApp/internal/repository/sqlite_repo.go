package repository

import (
	"context"
	"database/sql"
)

type sqliteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) TaskRepository {
	return &sqliteRepo{db: db}
}

func (r *sqliteRepo) SaveTask(ctx context.Context, task Task) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO items (user, product, amount, status) VALUES (?, ?, ?, 'pending')", task.User, task.Product, task.Amount)
	return err
}

func (r *sqliteRepo) GetTotalAmount(ctx context.Context, product string) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT SUM(amount) FROM items WHERE product = ?", product).Scan(&total)
	return total, err
}

func (r *sqliteRepo) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *sqliteRepo) GetPendingTasksCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM items WHERE status = 'pending'").Scan(&count)
	return count, err
}

func (r *sqliteRepo) GetNextPendingTask(ctx context.Context) (*Task, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, user, product, amount FROM items WHERE status = 'pending' LIMIT 1")
	var t Task
	err := row.Scan(&t.ID, &t.User, &t.Product, &t.Amount)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *sqliteRepo) MarkTaskDone(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE items SET status = 'done' WHERE id = ?", id)
	return err
}
