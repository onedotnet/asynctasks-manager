package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/onedotnet/asynctasks/database"
	"github.com/onedotnet/asynctasks/taskmanager"
	"github.com/spf13/cobra"
)

func makeatask() {
	// Code

	roopTask := taskmanager.TaskRoop{
		Source: "http://example.com",
		Target: "http://example.com",
		Output: "http://example.com",
	}

	task := taskmanager.Task{
		Name:     "Test Task",
		Status:   "Pending",
		Payload:  database.JSONB{"payload": roopTask},
		TaskType: taskmanager.TASK_TYPE_ROOP,
	}
	fmt.Println(task)

	node, err := taskmanager.GetAvaliableTaskNode("roop")
	if err != nil {
		fmt.Println(err)
		return
	}
	if node != nil {
		fmt.Println(node.Name)
	}

	q := taskmanager.DefaultQueueProvider
	j, _ := json.Marshal(task)
	q.PublishTo(node.Name, j)

}

var makeataskCmd = &cobra.Command{
	Use:   "newtask",
	Short: "Make a task",
	Run: func(cmd *cobra.Command, args []string) {
		makeatask()
	},
}

func init() {
	rootCmd.AddCommand(makeataskCmd)
}
