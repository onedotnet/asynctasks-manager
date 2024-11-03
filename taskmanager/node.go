package taskmanager

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/onedotnet/asynctasks/database"
)

const (
	NODE_CREATING     = "creating"
	NODE_RUNNING      = "running"
	NODE_EXECUTING    = "executing"
	NODE_PAUSED       = "paused"
	NODE_STOPPED      = "stopped"
	NODE_ERROR        = "error"
	NODE_UNAVAILIABLE = "unavailable"
	NODE_OFFLINE      = "offline"
)

type TaskNode struct {
	ID           int64          `json:"id" gorm:"primary_key"`
	NodeID       uuid.UUID      `json:"node_id" gorm:"type:uuid;unique_index"`
	Name         string         `json:"name" gorm:"varchar(255)"`
	Status       string         `json:"status" gorm:"varchar(255)"`
	ErrorMessage string         `json:"error_message" gorm:"varchar(255)"`
	Avaliable    bool           `json:"avaliable" gorm:"default:true"`
	Capabilities pq.StringArray `json:"capabilities" gorm:"type:text[]"`
	FinishedTask int64          `json:"finished_task" gorm:"default:0"`
	CPUNum       int            `json:"cpu_num" gorm:"default:1"`
	CPUUsage     float64        `json:"cpu_usage" gorm:"default:0"`
	Memory       int64          `json:"memory" gorm:"default:0"`
	MemoryUsage  float64        `json:"memory_usage" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at" gorm:"default:now()"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"default:now()"`
}

func CreateTaskNode(tn *TaskNode) error {
	if tn.Exists() {
		return fmt.Errorf("node %s already exists", tn.NodeID.String())
	}
	tn.CreatedAt = time.Now()
	tn.UpdatedAt = time.Now()
	tn.Avaliable = true
	tn.Status = NODE_CREATING
	tn.FinishedTask = 0

	return database.DB().Create(tn).Error
}

func (tn *TaskNode) Exists() bool {
	var count int64
	database.DB().Model(&TaskNode{}).Where("node_id = ?", tn.NodeID).Count(&count)
	if count > 0 {
		var ntn TaskNode
		database.DB().Model(&TaskNode{}).Where("node_id = ?", tn.NodeID).First(&ntn)
		tn.ID = ntn.ID
		return true
	}
	return false
}

func (tn *TaskNode) Update() error {
	if !tn.Exists() {
		return CreateTaskNode(tn)
	}
	tn.UpdatedAt = time.Now()
	fmt.Println(tn)
	return database.DB().Save(tn).Error
}

func (tn *TaskNode) SetAvaliable(avaliable, save bool) error {
	tn.Avaliable = avaliable
	if save {
		return tn.Update()
	}
	return nil
}

func (tn *TaskNode) SetStatus(status string, save bool) error {
	tn.Status = status
	if save {
		return tn.Update()
	}
	return nil
}

func (tn *TaskNode) SetFinishedTask(finished int64, save bool) error {
	tn.FinishedTask = finished
	if save {
		return tn.Update()
	}
	return nil
}

func (tn *TaskNode) SetStatusRunning(save bool) error {
	return tn.SetStatus(NODE_RUNNING, save)
}

func (tn *TaskNode) SetStatusPaused(save bool) error {
	return tn.SetStatus(NODE_PAUSED, save)
}

func (tn *TaskNode) SetStatusStopped(save bool) error {
	return tn.SetStatus(NODE_STOPPED, save)
}

func (tn *TaskNode) SetStatusUnavailable(save bool) error {
	return tn.SetStatus(NODE_UNAVAILIABLE, save)
}

func (tn *TaskNode) SetNodeAvaliable(save bool) error {
	return tn.SetAvaliable(true, save)
}

func (tn *TaskNode) SetNodeUnavaliable(save bool) error {
	return tn.SetAvaliable(false, save)
}

func GetTaskNode(nodeID uuid.UUID) (*TaskNode, error) {
	var tn TaskNode
	err := database.DB().Where("node_id = ?", nodeID).First(&tn).Error
	return &tn, err
}

func GetAvaliableTaskNode(capability string) (*TaskNode, error) {
	var tn TaskNode
	err := database.DB().Where("status = ? AND avaliable = ? AND ? = ANY(capabilities)", NODE_RUNNING, true, capability).First(&tn).Error
	return &tn, err
}

func GetTaskNodeList() ([]TaskNode, error) {
	var tns []TaskNode
	err := database.DB().Find(&tns).Error
	return tns, err
}
