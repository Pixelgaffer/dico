package main

import (
	"fmt"
	"net"

	protos "github.com/Pixelgaffer/dico-proto"

	"github.com/golang/protobuf/proto"
)

func handleClient(conn net.Conn) {
	defer conn.Close()
	data := make([]byte, 4096)
	n, err := conn.Read(data)
	checkErr(err)
	fmt.Println("Decoding Handshake")
	fmt.Println(data[0:n])
	hs, err := protos.DecodeUnknownMessage(data[0:n])
	checkErr(err)
	connection := &Connection{}
	connection.init(*hs.(*protos.Handshake))
	connection.conn = &conn
	connectionsLock.Lock()
	connections = append(connections, connection)
	connectionsLock.Unlock()

	go connection.handle()

	doneCh := make(chan interface{})

	go func() {
		for {
			data := make([]byte, 4096)
			n, err := conn.Read(data)
			if err != nil {
				fmt.Println("connection", conn.LocalAddr(), "<->", conn.RemoteAddr(), "died.")
				connection.dead = true
				return
			}
			fmt.Println("Decoding Packet")
			fmt.Println(data[0:n])
			msg, err := protos.DecodeUnknownMessage(data[0:n])
			if err != nil {
				connection.dead = true
				conn.Close()
				fmt.Println(err)
				return
			}
			connection.recv <- msg
		}
	}()

	for {
		select {
		case <-doneCh:
			return
		case msg := <-connection.send:
			wrapped := protos.WrapMessage(msg)
			data, err := proto.Marshal(wrapped)
			checkErr(err)
			_, err = conn.Write(data)
			if err != nil {
				fmt.Println("connection", conn.LocalAddr(), "<->", conn.RemoteAddr(), "died.")
				connection.dead = true
				return
			}
		}
	}
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

func listen() {
	fmt.Println("Listening on :7778...")
	ln, err := net.Listen("tcp", ":7778")
	checkErr(err)
	for {
		conn, err := ln.Accept()
		fmt.Println("new connection:", conn.LocalAddr(), "<->", conn.RemoteAddr())
		checkErr(err)
		go handleClient(conn)
	}
}
