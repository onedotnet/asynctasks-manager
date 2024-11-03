package taskmanager

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/onedotnet/asynctasks/database"
	"gorm.io/gorm"
)

/*
pending:
The task has been created but not yet picked up by a worker for processing.
Initial status after inserting the task into the PostgreSQL database and RabbitMQ queue.

in-progress:
The task has been assigned to a worker and is currently being processed.
This status is set as soon as the worker starts working on the task.

completed:
The task has been successfully processed and completed.
No further action required unless you want to run post-completion hooks or notifications.

failed:
The task has failed due to an error.
This can be an interim status before a retry, or final if the retry limit has been reached.

retrying:
The task failed and is currently being retried. It helps distinguish between fresh failures and retry attempts.
Retry logic increments the retry count.

cancelled:
The task was intentionally cancelled by the system or an administrator.
This could happen when the task is no longer relevant or has been manually aborted.

expired:
The task has exceeded its time-to-live (TTL) or deadline and is no longer valid.
This can be used in scenarios where tasks are time-sensitive (e.g., batch processing).

delayed:
The task is scheduled for later execution.
Useful for deferring non-urgent tasks, e.g., processing emails, notifications, etc.

paused (optional):
The task has been temporarily halted (e.g., due to external dependencies or maintenance).
Can be useful in long-running tasks that need to be paused and resumed later.

dead-lettered:
The task has been moved to a dead-letter queue due to max retries, expiration, or some unrecoverable failure.
You can configure RabbitMQ to automatically route failed tasks here.
*/
const (
	TASK_PENDING       = "pending"
	TASK_INPROGRESS    = "inprogress"
	TASK_COMPLETED     = "completed"
	TASK_FAILED        = "failed"
	TASK_CANCELLED     = "cancelled"
	TASK_RETRYING      = "retrying"
	TASK_EXPIRED       = "expired"
	TASK_DELAYED       = "delayed"
	TASK_PAUSED        = "paused"
	TASK_DEAD_LETTERED = "dead_lettered"

	TASK_TYPE_ROOP    = "roop"
	TASK_TYPE_CARTOON = "cartoon"
	TASK_TYPE_VIDEO   = "video"
)

type Task struct {
	ID        int64          `json:"id" gorm:"primary_key"`
	MessageID uuid.UUID      `json:"message_id" gorm:"type:uuid;unique_index"`
	Name      string         `json:"name" gorm:"varchar(255)"`
	Status    string         `json:"status" gorm:"varchar(255)"`
	Errors    pq.StringArray `json:"status_messages" gorm:"type:text[]"`
	Payload   database.JSONB `json:"payload" gorm:"type:jsonb"`
	TaskType  string         `json:"task_type" gorm:"varchar(255)"`
	Retried   int            `json:"retried" gorm:"default:0"`
	MaxRetry  int            `json:"max_retry" gorm:"default:3"`
	Deadline  int64          `json:"deadline" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at" gorm:"default:now()"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"default:now()"`
}

type TaskRoop struct {
	Source        string `json:"source"`
	Target        string `json:"target"`
	Output        string `json:"output"`
	Duration      int    `json:"duration"`
	Successful    bool   `json:"successful"`
	ErrorMessage  string `json:"error_message"`
	OutputMessage string `json:"output_message"`
}

func CreateTask(t *Task) error {
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	err := database.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(t).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func CreateRoopTask(tn *TaskRoop) (*Task, error) {
	task := Task{
		Name:     "Roop Task",
		Status:   TASK_PENDING,
		Payload:  database.JSONB{tn},
		TaskType: TASK_TYPE_ROOP,
	}
	err := database.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&task).Error; err != nil {
			return err
		}
		return nil
	})
	return &task, err
}

func (t *Task) Update() error {
	var count int64
	database.DB().Model(&Task{}).Where("id = ?", t.ID).Count(&count)
	if count == 0 {
		return fmt.Errorf("task %d not exists", t.ID)
	}
	t.UpdatedAt = time.Now()
	err := database.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(t).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}
