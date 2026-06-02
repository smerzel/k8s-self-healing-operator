package main

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// התחברות ישירה לתור ה-Redis שהרמנו בקלאסטר
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis-service:6379",
	})

	ctx := context.Background()
	queueKey := "sunday_tasks_queue"

	log.Println("Worker is running and waiting for tasks...")

	// לולאה אינסופית - ה-Worker כל הזמן מחכה לעבודה
	for {
		// BLPop - ה-Worker עוצר כאן וממתין בסבלנות. 
		// ברגע שתיכנס משימה חדשה לתור, הוא ישלוף אותה מיד.
		result, err := rdb.BLPop(ctx, 0, queueKey).Result()
		if err != nil {
			log.Printf("Error pulling from queue: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		taskData := result[1]
		log.Printf("Worker pulled a task: %s", taskData)

		// סימולציה של עיבוד כבד (למשל פנייה למסד נתונים או חישוב)
		time.Sleep(2 * time.Second)
		log.Println("Task completed successfully!")
	}
}