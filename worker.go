package main

import "fmt"
import protos "github.com/Pixelgaffer/dico-proto"

type Worker struct {
	connection     *Connection
	taskStatusChan chan *protos.TaskStatus
}

func (w *Worker) consume() {
	fmt.Println("consume()", taskChan)
	for task := range taskChan {
		task.worker = w
		task.execute(w.connection)
		if task.failed {
			retryChan <- task
			fmt.Println("resubmitted", task.id)
		}
	}
	panic("taskChan closed")
}
