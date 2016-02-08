package main

import (
	protos "github.com/Pixelgaffer/dico-proto"
	log "github.com/Sirupsen/logrus"
)

// Worker is directly linked to a Connection
type Worker struct {
	connection     *Connection
	taskStatusChan chan *protos.TaskStatus
	taskResultChan chan *protos.TaskResult
}

func (w *Worker) consume() {
	log.WithField("name", w.connection.name()).Info("worker started consuming")
	var task *Task
	for {
		select {
		case <-w.connection.doneCh:
			log.WithField("name", w.connection.name()).Info("worker stopped consuming")
			return
		case task = <-taskChan:
		}
		task.worker = w
		task.execute(w.connection)
		if task.failed {
			retryChan <- task
			log.WithFields(log.Fields{
				"id":      task.id,
				"options": task.options,
			}).Info("resubmitted task")
		}
	}
}
