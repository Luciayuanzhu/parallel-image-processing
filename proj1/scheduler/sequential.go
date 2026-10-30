package scheduler

import (
	"encoding/json"
	"os"
)

func RunSequential(config Config) {
	var total stageTimes
	for _, task := range readTasks(config.DataDirs) {
		stages := processTask(task, 1)
		total.Load += stages.Load
		total.Filter += stages.Filter
		total.Save += stages.Save
	}
	// Normal runs only print the time from editor.go.
	if os.Getenv("PROJ1_PROFILE") == "1" {
		if err := json.NewEncoder(os.Stderr).Encode(total); err != nil {
			panic(err)
		}
	}
}
