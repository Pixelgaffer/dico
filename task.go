package main

import (
	"fmt"
	"math/rand"
	"time"

	protos "github.com/Pixelgaffer/dico-proto"
	"github.com/golang/protobuf/proto"
)

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
