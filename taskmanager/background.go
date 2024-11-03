package taskmanager

import (
	"github.com/jasonlvhit/gocron"
	"github.com/onedotnet/asynctasks/database"
)

func Every10MinutesTask() {
	sql := `UPDATE task_nodes SET status = 'offline', 
	        avaliable = false
		WHERE updated_at < NOW() - INTERVAL '5 minutes';`
	database.DB().Exec(sql)
}

func StartBackGroundServices() {
	gocron.Every(10).Minutes().Do(Every10MinutesTask)
	gocron.Start()
}
