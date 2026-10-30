package scheduler

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"proj1/png"
	"strings"
	"sync"
	"time"
)

type imageTask struct {
	InPath  string   `json:"inPath"`
	OutPath string   `json:"outPath"`
	Effects []string `json:"effects"`
}

// Build the task list before starting workers.
func readTasks(dataDirs string) []imageTask {
	file, err := os.Open("../data/effects.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var tasks []imageTask
	for {
		var task imageTask
		err := decoder.Decode(&task)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		for _, dir := range strings.Split(dataDirs, "+") {
			tasks = append(tasks, imageTask{
				InPath:  filepath.Join("../data/in", dir, task.InPath),
				OutPath: filepath.Join("../data/out", dir+"_"+task.OutPath),
				Effects: task.Effects,
			})
		}
	}
	if err := os.MkdirAll("../data/out", 0755); err != nil {
		panic(err)
	}
	return tasks
}

// Measure load, filters, and save for the sequential Amdahl calculations.
type stageTimes struct {
	Load   float64 `json:"load"`
	Filter float64 `json:"filter"`
	Save   float64 `json:"save"`
}

func processTask(task imageTask, workers int) stageTimes {
	start := time.Now()
	img, err := png.Load(task.InPath)
	if err != nil {
		panic(err)
	}
	loaded := time.Now()
	applyEffects(img, task.Effects, workers)
	filtered := time.Now()
	if err := img.Save(task.OutPath); err != nil {
		panic(err)
	}
	return stageTimes{loaded.Sub(start).Seconds(), filtered.Sub(loaded).Seconds(), time.Since(filtered).Seconds()}
}

func applyEffects(img *png.Image, effects []string, workers int) {
	height := img.Bounds.Dy()
	if workers > height {
		workers = height
	}
	for _, effect := range effects {
		if workers <= 1 {
			img.ApplyRows(effect, img.Bounds.Min.Y, img.Bounds.Max.Y)
		} else {
			var wg sync.WaitGroup
			for id := 0; id < workers; id++ {
				start := img.Bounds.Min.Y + id*height/workers
				end := img.Bounds.Min.Y + (id+1)*height/workers
				wg.Add(1)
				go func(first, last int) {
					defer wg.Done()
					img.ApplyRows(effect, first, last)
				}(start, end)
			}
			// The next effect needs every row of this effect's result.
			wg.Wait()
		}
		img.Swap()
	}
}
