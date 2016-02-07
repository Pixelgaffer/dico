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
	fmt.Println("consume()", taskChan)
	for task := range taskChan {
		task.worker = w
		task.execute(w.connection)
		if task.failed {
			retryChan <- task
			fmt.Println("resubmitted", task.id)
		}
		if w.connection.dead {
			fmt.Println("worker", w.name(), "stopped consuming")
			return
		}
	}
	panic("taskChan closed")
}

func (w *Worker) name() string {
	conn := *w.connection.conn
	return fmt.Sprintf("%v [%v]", w.connection.handshake.GetName(), conn.LocalAddr())
}
