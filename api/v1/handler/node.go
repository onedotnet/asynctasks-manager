package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/onedotnet/asynctasks/taskmanager"
)

func NodeKeepAlive(c *gin.Context) {
	var node taskmanager.TaskNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if node.Exists() {
		node.Update()
	} else {
		taskmanager.CreateTaskNode(&node)
	}

	c.JSON(200, gin.H{"node": node})
}
