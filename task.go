package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Task struct {
	options string
	id      int
	retries int
	failed  bool
	worker  *Worker
	wg      *sync.WaitGroup
}

func (t *Task) execute() {
	t.failed = false
	fmt.Println("executing task", t.id, t.options)
	for i := 0; i < rand.Intn(10000)+10000; i++ {
		time.Sleep(time.Millisecond)
	}
	fmt.Println("finished executing task", t.id)
	if rand.Intn(100) <= 5 {
		t.failed = true
		return
	}

	t.wg.Done()
}
