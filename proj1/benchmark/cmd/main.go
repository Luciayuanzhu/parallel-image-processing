// This small driver measures Schedule without including compilation time.
package main

import (
	"encoding/json"
	"os"
	"proj1/scheduler"
	"strconv"
	"time"
)

func main() {
	threads, err := strconv.Atoi(os.Args[3])
	if err != nil {
		panic(err)
	}
	config := scheduler.Config{DataDirs: os.Args[1], Mode: os.Args[2], ThreadCount: threads}
	start := time.Now()
	scheduler.Schedule(config)
	json.NewEncoder(os.Stdout).Encode(map[string]float64{"total": time.Since(start).Seconds()})
}
