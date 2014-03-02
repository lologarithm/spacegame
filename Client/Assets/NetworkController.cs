using UnityEngine;
using System;
using System.Collections;
using System.Net;
using System.Net.Sockets;
using System.Collections.Generic;

public class NetworkController : MonoBehaviour {
	Socket sending_socket = new Socket(AddressFamily.InterNetwork, SocketType.Dgram, ProtocolType.Udp);
	IPAddress send_to_address;
	IPEndPoint sending_end_point;

	// Caching network state
	private byte[] buff = new byte[1024];
	private List<byte> stored_bytes = new List<byte>();
	private NetMessage current_message = null;

	// Use this for initialization
	void Start () {
		this.send_to_address = IPAddress.Parse("192.168.1.35");
		this.sending_end_point =  new IPEndPoint(send_to_address, 24816);
		LoginMessage login_msg = new LoginMessage ("a", "");
		sending_socket.Connect (this.sending_end_point);
		sending_socket.Send (login_msg.MessageBytes());
		// 1. Fetch network!
		// Start Receive and a new Accept
		try {
			sending_socket.BeginReceive(this.buff, 0, this.buff.Length, SocketFlags.None, new AsyncCallback(ReceiveCallback), null);
		} catch (SocketException e) {
			// DO something
			System.Console.WriteLine(e.ToString());
		}

	}

	private void ReceiveCallback(IAsyncResult result)
	{
		//PlayerConnection connection = (PlayerConnection)result.AsyncState;
		try
		{
			int bytesRead = sending_socket.EndReceive(result);
			if ( bytesRead > 0 )
			{
				//0. Add buffer to all_bytes
				//1. if ( connection.all_bytes.Count > 0 ) - Read int off front (package_size) 
				//2. while ( connection.all_bytes.Count + bytesRead >= package_size )
				//3. add buffer to all_bytes and then queue a message, delete bytes from all_bytes
				
				byte[] subset_bytes = new byte[bytesRead];
				Array.Copy(this.buff, 0, subset_bytes, 0, bytesRead);
				stored_bytes.AddRange(subset_bytes);
				ProcessBytes(stored_bytes.ToArray());
				sending_socket.BeginReceive(this.buff, 0, buff.Length, SocketFlags.None, new AsyncCallback(ReceiveCallback), null);
			}
			else 
				CloseConnection();
		}
		catch (SocketException exc)
		{
			CloseConnection();
			Console.WriteLine("Socket exception: " + exc.SocketErrorCode);
		}
		catch (Exception exc)
		{
			CloseConnection();
			Console.WriteLine("Exception: " + exc);
		}
	}
	
	private void ProcessBytes(byte[] input_bytes)
	{
		if (this.current_message != null) {
			NetMessage nMsg = NetMessage.fromBytes(input_bytes);
			if (nMsg != null) {
				// Check for full content available. If so, time to add this to the processing queue.
			} else {
				// Leave this as the this.current_message
			}
		} else {
			if (this.current_message.loadContent(input_bytes) ) {
				// We succeeded!
			} else {
				// We need to wait until later to finish loading!
			}
		}

	}

	// Update is called once per frame
	void Update () {
		// Send updates to each object.
	}

	void CloseConnection() {
		sending_socket.Send (new byte[]{255, 0, 0, 0, 0, 0, 0});
		sending_socket.Close();
	}

	void OnApplicationQuit() {
		CloseConnection ();
	}
}