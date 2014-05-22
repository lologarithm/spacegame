package server

import (
	"fmt"
)

func (s *Server) sendMessages() {
	for {
		msg := <-s.outToNetwork
		if msg.frame.message_type == 255 {
		}
		if n, err := s.conn.WriteToUDP(msg.rawBytes, msg.destination.clientAddress); err != nil {
			fmt.Println("Error: ", err, " Bytes Written: ", n)
		}
	}
}
