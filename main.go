package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/evalphobia/logrus_sentry"
)

var taskChan chan *Task
var retryChan chan *Task

const sentryDsn = "http://7e42960b144a40e39929367c1dd298c4:f14e11166c414e7baa61cc4aec5bcdf4@sentry.thuermchen.com/5" // TODO: https sentry

func init() {
	log.SetFormatter(&TextFormatter{})

	hook, err := logrus_sentry.NewSentryHook(sentryDsn, []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
	})

	if err == nil {
		log.AddHook(hook)
	} else {
		log.Error(err)
	}
}

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
}
