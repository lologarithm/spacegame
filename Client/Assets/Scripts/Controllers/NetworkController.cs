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

	private Queue<NetMessage> message_queue = new Queue<NetMessage>();

	// Use this for initialization
	void Start () {
		this.stored_bytes = new List<byte> ();
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
		int bytesRead = 0;
		try
		{
			bytesRead = sending_socket.EndReceive(result);
		}
		catch (SocketException exc)
		{
			CloseConnection();
			Debug.Log ("Socket exception: " + exc.SocketErrorCode);
			//			Console.WriteLine("Socket exception: " + exc.SocketErrorCode);
		}
		catch (Exception exc)
		{
			CloseConnection();
			Debug.Log ("Exception: " + exc);
			//			Console.WriteLine("Exception: " + exc);
		}

		if ( bytesRead > 0 )
		{
			//0. Add buffer to all_bytes
			//1. if ( connection.all_bytes.Count > 0 ) - Read int off front (package_size) 
			//2. while ( connection.all_bytes.Count + bytesRead >= package_size )
			//3. add buffer to all_bytes and then queue a message, delete bytes from all_bytes
			
			byte[] subset_bytes = new byte[bytesRead];
			Array.Copy(this.buff, 0, subset_bytes, 0, bytesRead);
			this.stored_bytes.AddRange(subset_bytes);
			ProcessBytes();
			sending_socket.BeginReceive(this.buff, 0, buff.Length, SocketFlags.None, new AsyncCallback(ReceiveCallback), null);
		}
		else 
			CloseConnection();
	}
	
	private void ProcessBytes()
	{
		byte[] input_bytes = this.stored_bytes.ToArray ();
		if (this.current_message == null) {
			NetMessage nMsg = NetMessage.fromBytes(input_bytes);
			if (nMsg != null) {
				// Check for full content available. If so, time to add this to the processing queue.
				if (nMsg.full_content.Length == nMsg.content_length + NetMessage.DEFAULT_FRAME_LEN) {
					stored_bytes.RemoveRange(0, nMsg.full_content.Length);
					this.message_queue.Enqueue(nMsg);
					// If we have enough bytes to start a new message we call ProcessBytes again.
					if (input_bytes.Length - nMsg.full_content.Length > NetMessage.DEFAULT_FRAME_LEN) {
						ProcessBytes();
					}
				}
			} else {
				this.current_message = nMsg;
				this.stored_bytes.RemoveRange(0, NetMessage.DEFAULT_FRAME_LEN);
				// Leave this as the this.current_message
			}
		} else {
			if (this.current_message.loadContent(input_bytes) ) {
				// We have enough bytes!
				stored_bytes.RemoveRange(0, this.current_message.content_length);
				this.message_queue.Enqueue(this.current_message);
				this.current_message = null;
			}
		}
		// We need to wait until later to finish loading!
	}

	// Update is called once per frame
	void Update () {
		int loops = this.message_queue.Count;
		for ( int i = 0; i <loops; i++ ){
			NetMessage msg = this.message_queue.Dequeue();
			switch ( msg.message_type ) {
			case NetMessage.ECHO:
				break;
			case NetMessage.LOGINSUCCESS:
				break;
			case NetMessage.LOGINFAIL:
				break;
			case NetMessage.PHYSICS:
				Debug.Log ("PHYSICS!");
				break;
			}
		}
		// Read from message queue and process!
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