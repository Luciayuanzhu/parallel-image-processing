package scheduler

func RunParallelSlices(config Config) {
	queue := taskQueue{tasks: readTasks(config.DataDirs)}
	for {
		task, ok := queue.pop()
		if !ok {
			return
		}
		// Finish this whole image before taking the next one.
		processTask(task, config.ThreadCount)
	}
}
