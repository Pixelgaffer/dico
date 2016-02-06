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

type Connection struct {
	name      string
	worker    *Worker
	conn      *net.Conn
	handshake protos.Handshake
	send      chan proto.Message
	recv      chan proto.Message
	dead      bool
}

func (c *Connection) init(handshake protos.Handshake) {
	c.send = make(chan proto.Message, 100) // TODO
	c.recv = make(chan proto.Message)
	c.handshake = handshake
	c.name = handshake.GetName()
	if handshake.GetRunsTasks() {
		c.worker = &Worker{}
	}
}

func (c *Connection) handle() {
	for {
		msg := <-c.recv
		switch v := msg.(type) {
		case *protos.SubmitTask:
			fmt.Println("taskSubmit", v)
		default:
			fmt.Println(proto.MessageName(msg), v)
		}
	}
}

func workers() (c chan *Worker) {
	go func() {
		for _, conn := range connections {
			if !conn.dead && conn.handshake.GetRunsTasks() {
				c <- conn.worker
			}
		}
	}()
	return c
}

func managers() (c chan *Connection) {
	go func() {
		for _, conn := range connections {
			if !conn.dead && conn.handshake.GetRecievesStats() {
				c <- conn
			}
		}
	}()
	return c
}
