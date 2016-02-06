package main

import "fmt"

type Worker struct {
}

func (w *Worker) consume() {
	fmt.Println("consume()", taskChan)
	for task := range taskChan {
		task.worker = w
		task.execute()
		if task.failed {
			retryChan <- task
			fmt.Println("resubmitted", task.id)
		}
	}
	panic("taskChan closed")
}
