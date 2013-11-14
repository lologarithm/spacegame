package main

import "testing"
import "bytes"
import "encoding/binary"
import "fmt"
import "net"
import "time"

func BenchmarkEcho(t *testing.B) {
	//exit := make(chan int, 1)
	//incoming_requests := make(chan Message, 200)
	//outgoing_player := make(chan Message, 200)
	//go RunServer(exit, incoming_requests, outgoing_player)
	//go ManageRequests(exit, incoming_requests, outgoing_player)
	time.Sleep(1 * time.Second)
	num_conn := 1000
	conns := [1000]*net.UDPConn{}
	ra, err := net.ResolveUDPAddr("udp", "localhost:24816")
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	for i := 0; i < num_conn; i++ {
		con, err := net.DialUDP("udp", nil, ra)
		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}
		conns[i] = con
		message_bytes := new(bytes.Buffer)
		message_bytes.WriteByte(1)
		binary.Write(message_bytes, binary.LittleEndian, int32(i))
		binary.Write(message_bytes, binary.LittleEndian, int32(1))
		message_bytes.WriteByte(97)
		con.Write(message_bytes.Bytes())
		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}
	}
	fmt.Println("Connections Complete")
	original_message := "This is a test message!"
	message_bytes := []byte(original_message)
	var msg_len = make([]byte, 4)
	binary.LittleEndian.PutUint32(msg_len, uint32(len(message_bytes)))
	output_message := append(append([]byte{0, 0, 0, 0, 0}, msg_len...), message_bytes...)
	var buf [512]byte
	//count := 0

	for i := 0; i < num_conn*500; i++ {
		var v = i % num_conn
		con := conns[v]
		_, err := con.Write(output_message)
		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}
		if n == 0 {
			t.FailNow()
		}
		//var c_len int32
		//binary.Read(bytes.NewBuffer(buf[5:9]), binary.LittleEndian, &c_len)
		//fmt.Println("Mes")
	}
	//exit <- 1
}

func TestLogin(t *testing.T) {
	exit := make(chan int, 1)
	incoming_requests := make(chan GameMessage, 200)
	outgoing_player := make(chan NetMessage, 200)
	go RunServer(exit, incoming_requests, outgoing_player)
	go ManageRequests(exit, incoming_requests, outgoing_player)
	time.Sleep(1 * time.Second)
	ra, err := net.ResolveUDPAddr("udp", "localhost:24816")
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	conn, err := net.DialUDP("udp", nil, ra)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	fmt.Println("Connection Complete")
	message_bytes := new(bytes.Buffer)
	message_bytes.WriteByte(1)
	binary.Write(message_bytes, binary.LittleEndian, int32(0))
	binary.Write(message_bytes, binary.LittleEndian, int32(1))
	message_bytes.WriteByte(97)
	conn.Write(message_bytes.Bytes())
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	buf := make([]byte, 1024)
	for i := 0; i < 10; i++ {
		n, err := conn.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			t.FailNow()
		}
		fmt.Println("Message recieved in test client: ", buf[0:n])
	}
	conn.Write([]byte{255, 0, 0, 0, 0, 0, 0, 0, 0})
	conn.Close()
}
