package main

import (
	"github.com/Mrmaxmeier/strgen"
	protos "github.com/Pixelgaffer/dico-proto"
	log "github.com/Sirupsen/logrus"
)

func generateTasks(str string) {
	g := &strgen.Generator{Source: str}
	log.WithFields(log.Fields{
		"amount":  g.Amount,
		"options": g.Source,
	}).Info("will generate tasks")
	if g.Err != nil {
		log.Error(g.Err)
		return
	}
	err := g.Configure()
	if err != nil {
		log.Error(err)
		return
	}
	go g.Generate()
	for g.Alive() {
		str, err := g.Next()
		if err != nil {
			log.Error(err)
			break
		}
		task := new(Task)
		task.id = getNextTaskID()
		task.options = str
		log.WithField("task", task).Info("generated task")
		task.reportStatus(protos.TaskStatus_REGISTERED)
		taskChan <- task
	}
	log.WithFields(log.Fields{
		"options": g.Source,
		"alive":   g.Alive(),
		"amount":  g.Amount,
		"left":    g.Left,
		"err":     g.Err,
	}).Info("generating tasks stopped")
}
