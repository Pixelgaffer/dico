package main

var taskChan chan *Task
var retryChan chan *Task
var currTaskId int64

func main() {
	taskChan = make(chan *Task)
	retryChan = make(chan *Task, 10) // TODO

	go func() {
		for task := range retryChan {
			task.retries++
			task.failed = false
			taskChan <- task
		}
	}()

	listen()

	//generateTasks("\\[0..2..8]", taskChan)
	//generateTasks("\\(one|two|3) \\(eins|zwei|drei)")
}
