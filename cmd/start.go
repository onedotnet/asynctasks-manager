package cmd

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"

	"github.com/gin-gonic/gin"
	"github.com/onedotnet/asynctasks/api/v1/handler"
	"github.com/onedotnet/asynctasks/config"
	"github.com/spf13/cobra"
)

func start() {
	r := gin.Default()
	r.Use(gin.Recovery(), cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"*"},
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Use(static.Serve("/uploads", static.LocalFile(config.AppConfig.StaticPath, true)))

	groupName := fmt.Sprintf("%s/%s", config.AppConfig.APIMagicPath, "api/v1")
	handler.Init(r.Group(groupName))

	addr := fmt.Sprintf("%s:%d", config.AppConfig.ListenHost, config.AppConfig.ListenPort)
	r.Run(addr)

}

var startCMD = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func init() {
	rootCmd.AddCommand(startCMD)
}
