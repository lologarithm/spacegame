package main

import (
	"fmt"
)

func (s *Server) sendMessages() {
	for {
		msg := <-s.outgoing_player
		if msg.frame.message_type == 255 {
		}
		if n, err := s.conn.WriteToUDP(msg.raw_bytes, msg.destination.client_address); err != nil {
			fmt.Println("Error: ", err, " Bytes Written: ", n)
		}
		fmt.Println("Sent")
	}
}
