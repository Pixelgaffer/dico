package main

import (
	"fmt"
	"net"
	"sync"

	protos "github.com/Pixelgaffer/dico-proto"
	"github.com/golang/protobuf/proto"
)

var connections []*Connection
var connectionsLock sync.Mutex

// Connection holds client-specific data
type Connection struct {
	name      string
	worker    *Worker
	conn      *net.Conn
	handshake protos.Handshake
	send      chan proto.Message
	recv      chan proto.Message
	doneCh    chan interface{}
}

func (c *Connection) init(handshake protos.Handshake) {
	fmt.Println("new Connection: runs_tasks", handshake.GetRunsTasks(), "manages_tasks", handshake.GetManagesTasks())
	c.send = make(chan proto.Message, 10) // TODO
	fmt.Println("c.send is:", c.send)
	c.recv = make(chan proto.Message)
	c.doneCh = make(chan interface{})
	c.handshake = handshake
	c.name = handshake.GetName()
	if handshake.GetRunsTasks() {
		c.worker = &Worker{
			connection:     c,
			taskStatusChan: make(chan *protos.TaskStatus),
			taskResultChan: make(chan *protos.TaskResult),
		}
		go c.worker.consume()
		fmt.Println("new worker consuming:", c.worker)
	}
}

func (c *Connection) handle() {
	var msg proto.Message
	for {
		select {
		case <-c.doneCh:
			return
		case msg = <-c.recv:
		}
		fmt.Println(msg)
		switch v := msg.(type) {
		case *protos.SubmitTask:
			fmt.Println("taskSubmit", v)
			if v.GetMulti() {
				generateTasks(v.GetOptions())
			} else {
				t := &Task{}
				t.id = getNextTaskID()
				t.options = v.GetOptions()
				taskChan <- t
			}
		case *protos.TaskStatus:
			c.worker.taskStatusChan <- v
		case *protos.TaskResult:
			c.worker.taskResultChan <- v
		default:
			fmt.Println(proto.MessageName(msg), v)
		}
	}
}

func (c *Connection) alive() bool {
	select {
	case <-c.doneCh:
		return false
	default:
		return true
	}
}

func (c *Connection) kill() {
	if c.alive() {
		conn := *c.conn
		fmt.Println("connection", conn.LocalAddr(), "<->", conn.RemoteAddr(), "died.")
		close(c.doneCh)
	} else {
		fmt.Println(".kill on dead connection")
	}
}

func workers() chan *Worker {
	connectionsLock.Lock()
	c := make(chan *Worker, len(connections))
	go func() {
		for _, conn := range connections {
			if conn.alive() && conn.handshake.GetRunsTasks() {
				c <- conn.worker
			}
		}
		close(c)
		connectionsLock.Unlock()
	}()
	return c
}

func managers() chan *Connection {
	connectionsLock.Lock()
	c := make(chan *Connection, len(connections))
	go func() {
		for _, conn := range connections {
			if conn.alive() && conn.handshake.GetRecievesStats() {
				c <- conn
			}
		}
		close(c)
		connectionsLock.Unlock()
	}()
	return c
}
