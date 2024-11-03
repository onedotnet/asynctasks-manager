package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "asynctasks",
	Short: "OneDtoNet AsyncTask Manager",
	Long:  `OneDotNet AsyncTask Manager is a tool to manage async tasks.`,
}

func initConfig() {

}

func init() {
	cobra.OnInitialize(initConfig)
}

func Execute() error {
	return rootCmd.Execute()
}
