package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/onedotnet/asynctasks/taskmanager"
)

func UpdateTask(c *gin.Context) {
	var task taskmanager.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := task.Update(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"task": task})
}

func GetTask(c *gin.Context) {
	taskID := c.Param("id")
	uid, err := uuid.Parse(taskID)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	task, err := taskmanager.GetTaskByUUID(uid)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"task": task})
}
