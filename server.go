package main

import (
	"bufio"
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
	protodata := new(protos.Handshake)
	err = proto.Unmarshal(data[0:n], protodata)
	checkErr(err)

	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("connection", conn.LocalAddr(), "<->", conn.RemoteAddr(), "died.")
			return
		}
		conn.Write([]byte(msg + "\n"))
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
