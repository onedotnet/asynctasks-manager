package handler

import "github.com/gin-gonic/gin"

func Init(rg *gin.RouterGroup) {
	rg.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "OneDotNet AsyncTask Manager",
			"version": "1.0.0",
			"author":  "OneDotNet LTD",
			"host":    c.Request.Host,
		})
	})

	// node routes
	rg.POST("/node/keepalive", NodeKeepAlive)

	// image routes
	rg.POST("/image/task/upload", UploadTaskImage)
	rg.POST("/artifact/upload", UploadArtifactImage)

	// task routes
	rg.POST("/task/update", UpdateTask)
}
