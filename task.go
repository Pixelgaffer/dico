package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

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
	//DEBUG: //FIXME
	fmt.Println("orig:")
	fmt.Println(proto.Marshal(m))
	wrappd := protos.WrapMessage(m)
	fmt.Println("wrapped:")
	fmt.Println(proto.Marshal(wrappd))

	for mang := range managers() {
		fmt.Println("reporting status to", mang, mang.send)
		mang.send <- m
	}
	fmt.Println("reported status", typ)
}

func (t *Task) reportResult(result string) {
	m := &protos.TaskResult{
		Id:      proto.Int64(t.id),
		Options: proto.String(t.options),
		Data:    []byte(result),
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
	for i := 0; i < rand.Intn(10000)+10000; i++ {
		time.Sleep(time.Millisecond)
	}
	if rand.Intn(100) <= 5 {
		fmt.Println("task", t.id, "failed!")
		t.failed = true
		t.reportStatus(protos.TaskStatus_FAILED)
		return
	}
	t.reportStatus(protos.TaskStatus_FINISHED)
	t.reportResult("huehuehue fake results huehuehue")
	fmt.Println("finished executing task", t.id)
}

func getNextTaskID() int64 {
	return atomic.AddInt64(&currTaskID, 1) - 1
}
