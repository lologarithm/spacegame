package server

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
	conn             *net.UDPConn
	connections      map[string]*Client
	disconnectPlayer chan Client
	outToNetwork     chan NetMessage
	toGameManager    chan GameMessage
	inputBuffer      []byte
	encryptionKey    *rsa.PrivateKey
}

func (s *Server) handleMessage() {
	// TODO: Add timeout on read to check for stale connections and add new user connections.
	s.conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	n, addr, err := s.conn.ReadFromUDP(s.inputBuffer)

	if err != nil {
		return
	}
	addr_str := addr.String()
	if n == 0 {
		s.DisconnectConn(addr_str)
	}
	if _, ok := s.connections[addr_str]; !ok {
		fmt.Printf("New Connection: %v\n", addr_str)
		s.connections[addr_str] = &Client{clientAddress: addr, fromNetwork: make(chan []byte, 100), fromGameManager: make(chan GameMessage, 10)}
		go s.connections[addr_str].ProcessBytes(s.toGameManager, s.outToNetwork, s.disconnectPlayer)
	}
	s.connections[addr_str].fromNetwork <- s.inputBuffer[0:n]
}

func (s *Server) DisconnectConn(addr_str string) {
	close(s.connections[addr_str].fromNetwork)
	delete(s.connections, addr_str)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())
		os.Exit(1)
	}
}

func RunServer(exit chan int, toGameManager chan GameMessage, outgoingNetwork chan NetMessage) {
	udpAddr, err := net.ResolveUDPAddr("udp", port)
	checkError(err)
	fmt.Println("Now listening on port", port)

	var s Server
	s.connections = make(map[string]*Client, 512)
	s.inputBuffer = make([]byte, 1024)
	s.toGameManager = toGameManager
	s.outToNetwork = outgoingNetwork
	s.disconnectPlayer = make(chan Client, 512)
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
		case client_obj := <-s.disconnectPlayer:
			s.DisconnectConn(client_obj.clientAddress.String())
		default:
			s.handleMessage()
		}
	}
}
