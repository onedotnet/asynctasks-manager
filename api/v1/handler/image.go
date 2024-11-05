package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	"github.com/onedotnet/asynctasks/config"
	"github.com/onedotnet/asynctasks/database"
	"github.com/onedotnet/asynctasks/taskmanager"
)

// imageDecode decodes a base64 encoded image string and returns the decoded image,
// its format (either "jpg" or "png"), and an error if any occurred during decoding.
//
// Parameters:
//   - i: A base64 encoded image string with a data URI scheme prefix.
//
// Returns:
//   - image.Image: The decoded image.
//   - string: The format of the image ("jpg" or "png").
//   - error: An error if the image type is unsupported or if decoding fails.
func imageDecode(i string) (image.Image, string, error) {
	// Code
	coI := strings.Index(i, ",")
	raw := i[coI+1:]
	unbased, _ := base64.StdEncoding.DecodeString(raw)
	res := bytes.NewReader(unbased)
	switch strings.TrimSuffix(i[5:coI], ";base64") {
	case "image/jpeg":
		jpgI, err := jpeg.Decode(res)
		if err != nil {
			return nil, "", err
		}
		return jpgI, "jpg", nil

	case "image/png":
		pngI, err := png.Decode(res)
		if err != nil {
			return nil, "", err
		}
		return pngI, "png", nil
	}
	return nil, "", fmt.Errorf("unsupported image type")
}

// UploadTaskImage handles the uploading of an image for a task.
// It performs the following steps:
// 1. Retrieves an available task node.
// 2. Binds the incoming JSON payload to an image struct.
// 3. Decodes the base64 image string to an image object.
// 4. Creates a directory for today's date if it doesn't exist.
// 5. Generates a unique filename for the image and saves it to the directory.
// 6. Creates a new task with the image information and saves it to the database.
// 7. Publishes the task to the task queue.
// 8. Returns the created task in the response.
//
// Parameters:
// - c: The Gin context, which provides request and response handling.
//
// Responses:
// - 200: Successfully created the task and returns the task information.
// - 400: Bad request, returns an error message if the JSON binding or image decoding fails.
// - 500: Internal server error, returns an error message if any other step fails.
func UploadTaskImage(c *gin.Context) {
	// Code

	var img struct {
		Image        string `json:"image" binding:"required"`
		BaseImageUrl string `json:"base_image_url" binding:"required"`
		TaskType     string `json:"task_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&img); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	node, err := taskmanager.GetAvaliableTaskNode(img.TaskType)
	if err != nil {
		c.JSON(500, gin.H{"get avaliable task node error": err.Error()})
		return
	}

	if node == nil {
		c.JSON(500, gin.H{"node is nil error": "no available node"})
		return
	}

	imgI, ext, err := imageDecode(img.Image)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	todaypath := fmt.Sprintf("%s/%s", config.AppConfig.StaticPath, time.Now().Format("2006-01-02"))
	if _, err := os.Stat(todaypath); os.IsNotExist(err) {
		err = os.Mkdir(todaypath, os.ModePerm)
		if err != nil {
			c.JSON(500, gin.H{"error": "create folder error: \n" + err.Error()})
			return
		}
	}
	taskid, _ := uuid.NewUUID()
	filename := fmt.Sprintf("%s-src.%s", taskid.String(), ext)

	f, err := os.Create(fmt.Sprintf("%s/%s", todaypath, filename))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	defer f.Close()
	switch ext {
	case "jpg":
		if err = jpeg.Encode(f, imgI, nil); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	case "png":
		if err = png.Encode(f, imgI); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	task := taskmanager.Task{
		Name:      "Image Task",
		TaskType:  img.TaskType,
		Status:    "Pending",
		MessageID: taskid,
	}

	sourceUrl := fmt.Sprintf("%s/%s/%s", config.AppConfig.InstancePublicURL, todaypath, filename)
	task.Payload = database.JSONB{"payload": taskmanager.TaskRoop{
		Source:        sourceUrl,
		Target:        img.BaseImageUrl,
		OutputMessage: "",
	}}
	// Code

	err = taskmanager.CreateTask(&task)
	if err != nil {
		c.JSON(500, gin.H{"create task error": err.Error()})
		return
	}

	q := taskmanager.DefaultQueueProvider
	j, _ := json.Marshal(task)
	q.PublishTo(node.Name, j)

	c.JSON(200, gin.H{"task": task})

}

func UploadArtifactImage(c *gin.Context) {
	for h := range c.Request.Header {
		fmt.Println(h, " ", c.Request.Header.Get(h))
	}
	todaypath := fmt.Sprintf("%s/artifacts/%s", config.AppConfig.StaticPath, time.Now().Format("2006-01-02"))
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err = c.SaveUploadedFile(file, fmt.Sprintf("%s/%s", todaypath, file.Filename))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"url": fmt.Sprintf("%s/%s/%s", config.AppConfig.InstancePublicURL, todaypath, file.Filename)})

}
