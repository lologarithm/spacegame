package main

import (
	"crypto/rsa"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	port string = ":24816"
)

type Server struct {
	conn              *net.UDPConn
	connections       map[string]*Client
	disconnect_player chan Client
	outgoing_player   chan NetMessage
	incoming_requests chan GameMessage
	input_buffer      []byte
	encryption_key    *rsa.PrivateKey
}

func (s *Server) handleMessage() {
	// TODO: Add timeout on read to check for stale connections and add new user connections.
	s.conn.SetReadDeadline(time.Now().Add(time.Second))
	n, addr, err := s.conn.ReadFromUDP(s.input_buffer)
	if err != nil {
		return
	}
	addr_str := addr.String()
	if n == 0 {
		s.DisconnectConn(addr_str)
	}
	if _, ok := s.connections[addr_str]; !ok {
		s.connections[addr_str] = &Client{client_address: addr, incoming_bytes: make(chan []byte, 100)}
		go s.connections[addr_str].ProcessBytes(s.incoming_requests, s.outgoing_player, s.disconnect_player)
	}
	s.connections[addr_str].incoming_bytes <- s.input_buffer[0:n]
}

func (s *Server) DisconnectConn(addr_str string) {
	fmt.Println("Disconnect from: ", addr_str)
	close(s.connections[addr_str].incoming_bytes)
	delete(s.connections, addr_str)
}

func (s *Server) sendMessages() {
	for {
		msg := <-s.outgoing_player
		if msg.frame.message_type == 255 {
		}
		if n, err := s.conn.WriteToUDP(msg.raw_bytes, msg.destination.client_address); err != nil {
			fmt.Println("Error: ", err, " Bytes Written: ", n)
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())
		os.Exit(1)
	}
}

func RunServer(exit chan int, requests chan GameMessage, outgoing_player chan NetMessage) {
	udpAddr, err := net.ResolveUDPAddr("udp", port)
	checkError(err)
	fmt.Println("Now listening on port", port)

	var s Server
	s.connections = make(map[string]*Client, 512)
	s.input_buffer = make([]byte, 1024)
	s.incoming_requests = requests
	s.outgoing_player = outgoing_player
	s.disconnect_player = make(chan Client, 512)
	s.conn, err = net.ListenUDP("udp", udpAddr)
	checkError(err)

	go s.sendMessages()
ML:
	for {
		select {
		case <-exit:
			fmt.Println("Killing Socket Server")
			s.conn.Close()
			break ML
		case client_obj := <-s.disconnect_player:
			s.DisconnectConn(client_obj.client_address.String())
		default:
			s.handleMessage()
		}
	}
}
