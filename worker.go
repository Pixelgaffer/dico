package main

import "fmt"
import protos "github.com/Pixelgaffer/dico-proto"

// Worker is directly linked to a Connection
type Worker struct {
	connection     *Connection
	taskStatusChan chan *protos.TaskStatus
	taskResultChan chan *protos.TaskResult
}

func (w *Worker) consume() {
	fmt.Println("worker", w.connection.name(), "started consuming")
	var task *Task
	for {
		select {
		case <-w.connection.doneCh:
			fmt.Println("worker", w.connection.name(), "stopped consuming")
			return
		case task = <-taskChan:
		}
		task.worker = w
		task.execute(w.connection)
		if task.failed {
			retryChan <- task
			fmt.Println("resubmitted", task.id)
		}
	}
}
