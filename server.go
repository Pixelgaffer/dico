package main

import (
	"fmt"
	"net"

	protos "github.com/Pixelgaffer/dico-proto"
	"github.com/golang/protobuf/proto"

	log "github.com/Sirupsen/logrus"
)

func handleClient(conn net.Conn) {
	defer conn.Close()
	buff, err := protos.ReadPacket(conn)
	checkErr(err)
	log.WithField("buff", buff).Debug("Decoding Handshake")
	hs, err := protos.DecodeUnknownMessage(buff)
	checkErr(err)
	connection := &Connection{conn: &conn}
	connection.init(*hs.(*protos.Handshake))
	connections.Lock()
	connections.all = append(connections.all, connection)
	connections.Unlock()
	gcConnections()

	go connection.handle()

	go func() {
		for {
			buff, err := protos.ReadPacket(conn)
			if err != nil {
				log.WithField("addr", conn.RemoteAddr()).Warn(err)
				connection.kill()
				return
			}
			log.WithField("buff", buff).Debug("Decoding Packet")
			msg, err := protos.DecodeUnknownMessage(buff)
			if err != nil {
				log.WithField("addr", conn.RemoteAddr()).Warn(err)
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
			log.WithFields(log.Fields{
				"addr":    conn.RemoteAddr(),
				"message": msg,
			}).Debug("sending data")
			wrapped := protos.WrapMessage(msg)
			data, err := proto.Marshal(wrapped)
			checkErr(err)
			err = protos.WritePacket(conn, data)
			if err != nil {
				log.WithField("addr", conn.RemoteAddr()).Warn(err)
				connection.kill()
			}
		}
	}
}

func checkErr(e error) {
	if e != nil {
		log.Warn(e)
	}
}

func listen(port int) {
	pstr := fmt.Sprintf(":%d", port)
	log.Info("Listening on " + pstr + "...")
	ln, err := net.Listen("tcp", pstr)
	checkErr(err)
	for {
		conn, err := ln.Accept()
		log.WithField("addr", conn.RemoteAddr()).Info("new connection")
		checkErr(err)
		go handleClient(conn)
	}
}
