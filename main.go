package main

import "sync"

func main() {
	listen()

	taskChan := make(chan *Task)
	retryChan := make(chan *Task, 10) // TODO
	for i := 0; i < 3; i++ {
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
