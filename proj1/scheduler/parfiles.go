package scheduler

import "sync"

func RunParallelFiles(config Config) {
	queue := taskQueue{tasks: readTasks(config.DataDirs)}
	workers := config.ThreadCount
	if workers < 1 {
		workers = 1
	}
	if workers > len(queue.tasks) {
		workers = len(queue.tasks)
	}
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				task, ok := queue.pop()
				if !ok {
					return
				}
				// The queue is already unlocked. Each worker owns its image.
				processTask(task, 1)
			}
		}()
	}
	wg.Wait()
}
