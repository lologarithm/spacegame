package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/binary"
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
	outgoing_player   chan Message
	incoming_requests chan Message
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

func ParseFrame(raw_bytes []byte) *MessageFrame {
	if len(raw_bytes) >= 9 {
		mf := new(MessageFrame)
		mf.message_type = raw_bytes[0]
		var v int32
		binary.Read(bytes.NewBuffer(raw_bytes[1:5]), binary.LittleEndian, &v)
		mf.from_user = v
		binary.Read(bytes.NewBuffer(raw_bytes[5:9]), binary.LittleEndian, &v)
		mf.content_length = v
		mf.frame_length = 9
		return mf
	}

	return nil
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

type Client struct {
	buffer         []byte
	client_address *net.UDPAddr
	incoming_bytes chan []byte
	user           *User
}

func (client *Client) ProcessBytes(to_client chan Message, outgoing_msg chan Message, disconnect_player chan Client) {
	for {
		if dem_bytes, ok := <-client.incoming_bytes; !ok {
			break
		} else {
			client.buffer = append(client.buffer, dem_bytes...)
			msg_frame := ParseFrame(client.buffer)
			if msg_frame != nil && int(msg_frame.frame_length+msg_frame.content_length) >= len(client.buffer) {
				msg_obj := client.parseMessage(msg_frame)
				msg_obj.destination = client
				if msg_obj.frame.message_type == 0 {
					outgoing_msg <- msg_obj
				} else {
					to_client <- msg_obj
					if msg_obj.frame.message_type == 255 {
						disconnect_player <- *client
						break
					}
				}
			}
		}
	}
}

func (client *Client) parseMessage(mf *MessageFrame) (m Message) {
	m.raw_bytes = client.buffer[0 : mf.frame_length+mf.content_length]
	m.frame = mf
	client.buffer = client.buffer[mf.frame_length+mf.content_length:]
	return
}

type Message struct {
	raw_bytes   []byte
	frame       *MessageFrame
	destination *Client
}

func (m *Message) Content() []byte {
	return m.raw_bytes[m.frame.frame_length : m.frame.frame_length+m.frame.content_length]
}

type MessageFrame struct {
	message_type   byte // 0: echo, 1: login_request 2: login success 3: login failure/logoff 4: physics update
	from_user      int32
	frame_length   int32
	content_length int32
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())
		os.Exit(1)
	}
}

func RunServer(exit chan int, requests chan Message, outgoing_player chan Message) {
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
