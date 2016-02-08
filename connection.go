package main

import (
	"fmt"
	"net"
	"sync"

	protos "github.com/Pixelgaffer/dico-proto"
	"github.com/golang/protobuf/proto"

	log "github.com/Sirupsen/logrus"
)

var connections []*Connection
var connectionsLock sync.Mutex

// Connection holds client-specific data
type Connection struct {
	worker    *Worker
	conn      *net.Conn
	handshake protos.Handshake
	send      chan proto.Message
	recv      chan proto.Message
	doneCh    chan interface{}
}

func (c *Connection) init(handshake protos.Handshake) {
	log.WithFields(log.Fields{
		"runs_tasks":     handshake.GetRunsTasks(),
		"manages_tasks":  handshake.GetManagesTasks(),
		"recieves_stats": handshake.GetRecievesStats(),
	}).Info("new connection")
	c.send = make(chan proto.Message, 10) // TODO
	c.recv = make(chan proto.Message)
	c.doneCh = make(chan interface{})
	c.handshake = handshake
	if handshake.GetRunsTasks() {
		c.worker = &Worker{
			connection:     c,
			taskStatusChan: make(chan *protos.TaskStatus),
			taskResultChan: make(chan *protos.TaskResult),
		}
		go c.worker.consume()
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
		log.WithField("msg", msg).Debug("new msg")
		switch v := msg.(type) {
		case *protos.SubmitTask:
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
		case *protos.SubmitCode:
			addJobType(v)
		default:
			log.WithFields(log.Fields{
				"type":    proto.MessageName(msg),
				"message": v,
			}).Error("connection.handle invalid type")
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
		log.WithField("conn", c.name()).Info("connection died.")
		close(c.doneCh)
	} else {
		log.Info(".kill on dead connection")
	}
}

func (c *Connection) name() string {
	conn := *c.conn
	if c.handshake.GetName() == "" {
		return fmt.Sprintf("[%v]", conn.RemoteAddr())
	}
	return fmt.Sprintf("%v [%v]", c.handshake.GetName(), conn.RemoteAddr())
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
