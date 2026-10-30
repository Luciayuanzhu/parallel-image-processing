package scheduler

import (
	"runtime"
	"sync/atomic"
)

// A test-and-set lock: Swap returns the value before setting it to 1.
// Only the worker that sees 0 has acquired the lock.
type tasLock struct{ state int32 }

func (lock *tasLock) Lock() {
	for atomic.SwapInt32(&lock.state, 1) != 0 {
		runtime.Gosched() // Let another goroutine run while we wait.
	}
}

func (lock *tasLock) Unlock() { atomic.StoreInt32(&lock.state, 0) }

type taskQueue struct {
	lock  tasLock
	tasks []imageTask
	next  int
}

func (queue *taskQueue) pop() (imageTask, bool) {
	queue.lock.Lock()
	defer queue.lock.Unlock()
	if queue.next == len(queue.tasks) {
		return imageTask{}, false
	}
	task := queue.tasks[queue.next]
	queue.next++
	return task, true
}
