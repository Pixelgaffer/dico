package main

import (
	"fmt"
	"sync"
	"sync/atomic"

	protos "github.com/Pixelgaffer/dico-proto"
	"github.com/golang/protobuf/proto"
)

var currTaskID int64
var taskIDLock sync.Mutex

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
	fmt.Println("executing task", t.id, t.options)
	t.reportStatus(protos.TaskStatus_REGISTERED)
	t.failed = false
	c.send <- &protos.DoTask{
		Id:      proto.Int64(t.id),
		Options: proto.String(t.options),
		JobType: proto.String("TODO"),
	}
	t.reportStatus(protos.TaskStatus_STARTED)
	for {
		select {
		case status := <-t.worker.taskStatusChan:
			switch status.GetType() {
			case protos.TaskStatus_FAILED:
				fmt.Println("task failed:", t.id)
				t.failed = true
				t.reportStatus(status.GetType())
				return
			case protos.TaskStatus_FINISHED:
				t.reportStatus(status.GetType())
			default:
				fmt.Println("invalid status.Type")
			}
		case result := <-t.worker.taskResultChan:
			fmt.Println("got task result")
			t.reportResult(result.Data)
			return
		}
	}
}

func getNextTaskID() int64 {
	return atomic.AddInt64(&currTaskID, 1) - 1
}
