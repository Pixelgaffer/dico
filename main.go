package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/bobziuchkovski/writ"

	"github.com/evalphobia/logrus_sentry"
)

var taskChan chan *Task
var retryChan chan *Task

const sentryDsn = "http://7e42960b144a40e39929367c1dd298c4:f14e11166c414e7baa61cc4aec5bcdf4@sentry.thuermchen.com/5" // TODO: https sentry

type Dicod struct {
	HelpFlag  bool `flag:"help" description:"Display this help message and exit"`
	Verbosity int  `flag:"v, verbose" description:"Display verbose output"`
	Port      int  `option:"p, port" default:"7778" description:"The port the server runs on"`
}

func init() {
	log.SetFormatter(&TextFormatter{})

	go func() {
		hook, err := logrus_sentry.NewSentryHook(sentryDsn, []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
		})

		if err == nil {
			log.AddHook(hook)
			log.Debug("added sentry hook")
		} else {
			log.Error(err)
		}
	}()
}

func main() {

	dicod := &Dicod{}
	cmd := writ.New("dicod", dicod)
	cmd.Help.Usage = "Usage: dicod [OPTION]..."
	cmd.Help.Header = "Distributes tasks, collects results."
	_, _, err := cmd.Decode(os.Args[1:])
	if err != nil || dicod.HelpFlag {
		cmd.ExitHelp(err)
	}
	if dicod.Verbosity > 0 {
		log.SetLevel(log.DebugLevel)
		log.WithField("verbosity", dicod.Verbosity).Debug("custom verbosity")
	}

	taskChan = make(chan *Task)
	retryChan = make(chan *Task, 10) // TODO

	go func() {
		for task := range retryChan {
			task.retries++
			task.failed = false
			taskChan <- task
		}
	}()

	listen(dicod.Port)
}
