using UnityEngine;
using System.Collections;
using System.Net;
using System.Net.Sockets;

public class NetworkController : MonoBehaviour {
	Socket sending_socket = new Socket(AddressFamily.InterNetwork, SocketType.Dgram, ProtocolType.Udp);
	IPAddress send_to_address;
	IPEndPoint sending_end_point;

	// Use this for initialization
	void Start () {
		this.send_to_address = IPAddress.Parse("192.168.1.35");
		this.sending_end_point =  new IPEndPoint(send_to_address, 24816);
		LoginMessage login_msg = new LoginMessage ("a", "");
		sending_socket.Connect (this.sending_end_point);
		sending_socket.Send (login_msg.MessageBytes());
	}
	
	// Update is called once per frame
	void Update () {
		// 1. Fetch network!
		// 2. Send updates to each object
	}

	void OnApplicationQuit() {
		sending_socket.Send (new byte[]{255, 0, 0, 0, 0, 0, 0, 0, 0});
	}
}
