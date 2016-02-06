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

var currTaskId int64
var taskIdLock sync.Mutex

type Task struct {
	options string
	id      int64
	retries int64
	failed  bool
	worker  *Worker
}

func (t *Task) execute(c *Connection) {
	t.failed = false
	c.send <- &protos.DoTask{
		Id:      proto.Int64(t.id),
		Options: proto.String(t.options),
		JobType: proto.String("TODO"),
	}
	fmt.Println("executing task", t.id, t.options)
	for i := 0; i < rand.Intn(10000)+10000; i++ {
		time.Sleep(time.Millisecond)
	}
	if rand.Intn(100) <= 5 {
		fmt.Println("task", t.id, "failed!")
		t.failed = true
		return
	}
	fmt.Println("finished executing task", t.id)
}

func getNextTaskID() int64 {
	return atomic.AddInt64(&currTaskId, 1) - 1
}
