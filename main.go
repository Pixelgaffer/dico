package main

import "sync"

const countdoku = 25

func main() {
	/*
		ui := UI{}
		defer ui.cleanup()
		ui.init()
		ui.block()
	*/

	taskChan := make(chan *Task)
	retryChan := make(chan *Task, countdoku*2)
	for i := 0; i < countdoku; i++ {
		worker := new(Worker)
		go worker.consume(taskChan, retryChan)
	}
	go func() {
		for task := range retryChan {
			task.retries++
			task.failed = false
			taskChan <- task
		}
	}()
	waitGroup := new(sync.WaitGroup)
	//generateTasks("\\[0..2..8]", taskChan, waitGroup)
	generateTasks("\\(one|two|3) \\(eins|zwei|drei)", taskChan, waitGroup)
	waitGroup.Wait()
	close(taskChan)
}
