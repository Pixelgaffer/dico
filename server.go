package main

import (
	"fmt"
	"net"

	protos "github.com/Pixelgaffer/dico-proto"

	"github.com/golang/protobuf/proto"
)

func handleClient(conn net.Conn) {
	defer conn.Close()
	buff, err := protos.ReadPacket(conn)
	checkErr(err)
	fmt.Println("Decoding Handshake")
	fmt.Println(buff)
	hs, err := protos.DecodeUnknownMessage(buff)
	checkErr(err)
	connection := &Connection{}
	connection.conn = &conn
	connection.init(*hs.(*protos.Handshake))
	connectionsLock.Lock()
	connections = append(connections, connection)
	connectionsLock.Unlock()

	go connection.handle()

	go func() {
		for {
			buff, err := protos.ReadPacket(conn)
			if err != nil {
				fmt.Println(err)
				connection.kill()
				return
			}
			fmt.Println("Decoding Packet")
			fmt.Println(buff)
			msg, err := protos.DecodeUnknownMessage(buff)
			if err != nil {
				fmt.Println(err)
				connection.kill()
				return
			}
			connection.recv <- msg
		}
	}()

	for {
		select {
		case <-connection.doneCh:
			return
		case msg := <-connection.send:
			fmt.Println("sending to", conn.RemoteAddr(), msg)
			wrapped := protos.WrapMessage(msg)
			data, err := proto.Marshal(wrapped)
			checkErr(err)
			err = protos.WritePacket(conn, data)
			if err != nil {
				fmt.Println(err)
				close(connection.doneCh)
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
