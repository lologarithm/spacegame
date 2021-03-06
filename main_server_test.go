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
		binary.Write(message_bytes, binary.LittleEndian, int16(1))
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
	fmt.Println("\nStarting Login Test")
	exit := make(chan int, 1)
	incomingRequests := make(chan GameMessage, 200)
	outgoingPlayer := make(chan NetMessage, 200)
	go RunServer(exit, incomingRequests, outgoingPlayer)
	go ManageRequests(exit, incomingRequests)
	time.Sleep(time.Millisecond * 50)
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
	//fmt.Println("Connection Complete")
	messageBytes := new(bytes.Buffer)
	messageBytes.WriteByte(1)
	binary.Write(messageBytes, binary.LittleEndian, uint16(0))
	binary.Write(messageBytes, binary.LittleEndian, uint16(3))
	messageBytes.WriteByte(97)
	messageBytes.WriteByte(58)
	messageBytes.WriteByte(97)
	conn.Write(messageBytes.Bytes())
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	buf := make([]byte, 512)
	n, err := conn.Read(buf[0:])
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	if n < 5 || buf[0] != 2 {
		t.FailNow()
	}
	conn.Write([]byte{255, 0, 0, 0, 0})
	conn.Close()
	fmt.Printf("TestLogin Test Pass.\n\n")
}

func TestSetThrust(t *testing.T) {
	//exit := make(chan int, 1)
	//incoming_requests := make(chan GameMessage, 200)
	//outgoing_player := make(chan NetMessage, 200)
	//go RunServer(exit, incoming_requests, outgoing_player)
	//go ManageRequests(exit, incoming_requests, outgoing_player)
	//time.Sleep(1 * time.Second)
	fmt.Printf("Testing SetThrustMessage.\n")
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
	buf := make([]byte, 1024)
	fmt.Println("Logging in ")
	message_bytes := new(bytes.Buffer)
	message_bytes.WriteByte(1)
	binary.Write(message_bytes, binary.LittleEndian, uint16(0))
	binary.Write(message_bytes, binary.LittleEndian, uint16(3))
	message_bytes.WriteByte(97)
	message_bytes.WriteByte(58)
	message_bytes.WriteByte(97)
	conn.Write(message_bytes.Bytes())
	n, err := conn.Read(buf[0:])
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	fmt.Printf("TEST: Received: %v\n", buf[0:n])

	message_bytes = new(bytes.Buffer)
	message_bytes.WriteByte(5)
	binary.Write(message_bytes, binary.LittleEndian, uint16(1))
	binary.Write(message_bytes, binary.LittleEndian, uint16(1))
	binary.Write(message_bytes, binary.LittleEndian, uint8(50))
	conn.Write(message_bytes.Bytes())

	HandleShipUpdateMessage(conn, t)
	HandleShipUpdateMessage(conn, t)
	HandleShipUpdateMessage(conn, t)
	HandleShipUpdateMessage(conn, t)
	HandleShipUpdateMessage(conn, t)

	// Do something here.
	conn.Write([]byte{255, 0, 0, 0, 0})
	conn.Close()
	fmt.Printf("TestSetThrustMessage Complete.\n\n")
	exit <- 1
}

func HandleShipUpdateMessage(conn *net.UDPConn, t *testing.T) []*Ship {
	buf := make([]byte, 1024)

	n, err := conn.Read(buf[0:])
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	fmt.Printf("TEST: Received: %v\n", buf[0:n])
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	msg_frame, _ := ParseFrame(buf[0:n])
	content := buf[msg_frame.frame_length : msg_frame.frame_length+msg_frame.content_length]
	ships := []*Ship{&Ship{RigidBody: RigidBody{Position: Vect2{0, 0}, Velocity: Vect2{0, 0}, Force: Vect2{0, 0}}}}

	for i := 0; i*36 < len(content); i += 1 {
		if len(ships) <= i {
			ships = append(ships, &Ship{RigidBody: RigidBody{Position: Vect2{0, 0}, Velocity: Vect2{0, 0}, Force: Vect2{0, 0}}})
		}
		ships[i].FromBytes(content[i*36 : (i+1)*36])
	}

	fmt.Printf("Ships: %v\n", ships)
	return ships
}
