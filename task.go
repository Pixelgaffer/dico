package main

import (
	"sync/atomic"

	protos "github.com/Pixelgaffer/dico-proto"
	"github.com/golang/protobuf/proto"

	log "github.com/Sirupsen/logrus"
)

var currTaskID int64

// Task is executed by an available worker
type Task struct {
	options string
	id      int64
	retries int64
	failed  bool
	worker  *Worker
}

func (t *Task) reportStatus(typ protos.TaskStatus_TaskStatusUpdate) {
	m := &protos.TaskStatus{
		Id:      proto.Int64(t.id),
		Options: proto.String(t.options),
		Retries: proto.Int64(t.retries),
	}
	m.Type = &typ
	if t.worker != nil {
		m.Worker = proto.String(t.worker.connection.name())
	}
	for mang := range managers() {
		mang.send <- m
	}
}

func (t *Task) reportResult(data []byte) {
	m := &protos.TaskResult{
		Id:      proto.Int64(t.id),
		Options: proto.String(t.options),
		Data:    data,
	}
	for mang := range managers() {
		mang.send <- m
	}
}

func (t *Task) execute(c *Connection) {
	log.WithFields(log.Fields{
		"id":      t.id,
		"options": t.options,
	}).Info("executing task")
	t.failed = false
	c.send <- &protos.DoTask{
		Id:      proto.Int64(t.id),
		Options: proto.String(t.options),
		Code:    proto.String("TODO"),
		JobType: proto.String("TODO"),
	}
	t.reportStatus(protos.TaskStatus_STARTED)
	for {
		select {
		case status := <-t.worker.taskStatusChan:
			switch status.GetType() {
			case protos.TaskStatus_FAILED:
				log.WithFields(log.Fields{
					"id":      t.id,
					"options": t.options,
				}).Info("task failed")
				t.failed = true
				t.reportStatus(status.GetType())
				return
			case protos.TaskStatus_FINISHED:
				t.reportStatus(status.GetType())
			default:
				log.WithField("status", status).Error("invalid status.Type")
			}
		case result := <-t.worker.taskResultChan:
			log.WithField("id", t.id).Info("got task result")
			t.reportResult(result.Data)
			return
		case <-t.worker.connection.doneCh:
			t.failed = true
			t.reportStatus(protos.TaskStatus_FAILED)
			return
		}
	}
}

func getNextTaskID() int64 {
	return atomic.AddInt64(&currTaskID, 1) - 1
}
