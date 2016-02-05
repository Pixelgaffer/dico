package main

import "fmt"

type Worker struct {
}

func (w *Worker) consume(taskChan, retryChan chan *Task) {
	for task := range taskChan {
		task.worker = w
		task.execute()
		if task.failed {
			retryChan <- task
			fmt.Println("resubmitted", task.id)
		}
	}
}
