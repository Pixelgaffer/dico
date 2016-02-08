package main

import (
	protos "github.com/Pixelgaffer/dico-proto"
	"github.com/Pixelgaffer/dicod/strgen"
	log "github.com/Sirupsen/logrus"
)

func generateTasks(str string) {
	c, a, err := strgen.GenerateStrings(str)
	log.WithFields(log.Fields{
		"amount":  a,
		"options": str,
	}).Info("will generate tasks")
	if err != nil {
		log.Error(err)
		return
	}
	for str := range c {
		task := new(Task)
		task.id = getNextTaskID()
		task.options = str
		log.WithField("task", task).Info("generated task")
		task.reportStatus(protos.TaskStatus_REGISTERED)
		taskChan <- task
	}
}
