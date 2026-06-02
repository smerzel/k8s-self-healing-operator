package repository

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

type redisRepo struct {
	client   *redis.Client
	queueKey string
}

// NewRedisRepo יוצר חיבור חדש לשרת ה-Redis בקלאסטר
func NewRedisRepo(addr string) TaskRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr, 
	})
	return &redisRepo{
		client:   rdb,
		queueKey: "sunday_tasks_queue", // השם של התור שבו נשמור את המשימות
	}
}

// PushTask מקבל משימה, הופך אותה ל-JSON ודוחף לסוף התור
func (r *redisRepo) PushTask(ctx context.Context, task Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	
	// LPush = Left Push. פעולה בסיסית ב-Redis שדוחפת פריט לתור
	return r.client.LPush(ctx, r.queueKey, data).Err()
}

// GetPendingTasksCount מחזיר את כמות המשימות שממתינות בתור
func (r *redisRepo) GetPendingTasksCount(ctx context.Context) (int64, error) {
	// LLen = List Length. קורא את אורך התור ישירות מהזיכרון
	return r.client.LLen(ctx, r.queueKey).Result()
}

func (r *redisRepo) Close() error {
	return r.client.Close()
}