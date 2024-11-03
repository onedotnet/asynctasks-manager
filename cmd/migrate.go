package cmd

import (
	"log/slog"

	"github.com/onedotnet/asynctasks/database"
	"github.com/onedotnet/asynctasks/taskmanager"
	"github.com/spf13/cobra"
)

var tableList = map[string]interface{}{
	"node": taskmanager.TaskNode{},
	"task": taskmanager.Task{},
}

func migrate() {
	slog.Info("Migrating database...")
	db := database.DB()
	for tablename, table := range tableList {
		slog.Info("Creating table ", "tablename=", tablename)
		db.AutoMigrate(table)

	}
	slog.Info("Database migration completed.")
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate the database",
	Run: func(cmd *cobra.Command, args []string) {
		migrate()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
