using System;
using System.IO;
using System.Text;


//ECHO         NetMessageType = 0
//LOGINREQUEST NetMessageType = 1
//LOGINSUCCESS NetMessageType = 2
//LOGINFAIL    NetMessageType = 3
//PHYSICS      NetMessageType = 4
//DISCONNECT   NetMessageType = 255

public class NetMessage {
	public byte message_type;
	public Int32 from_player;
	public Int16 content_length;
	public byte[] content;
	public byte[] full_content;


	public byte[] MessageBytes() {
		///byte[] byte_array = new byte[]
		MemoryStream stream = new MemoryStream();
		using (BinaryWriter writer = new BinaryWriter(stream))
		{
			writer.Write(this.message_type);
			writer.Write(from_player);
			writer.Write(content_length);
			writer.Write(content);
		}
		return stream.ToArray();
	}

	public byte[] Content() {
		byte[] content = null;
		Array.Copy (this.full_content, 9, content, 0, this.full_content.Length - 9);
		return content;
	}

	public static NetMessage fromBytes(byte[] bytes) {
		NetMessage newMsg = null;
		if (bytes.Length >= 9) {
			newMsg = new NetMessage ();
			newMsg.message_type = bytes[0];
			newMsg.from_player = BitConverter.ToInt32 (bytes, 1);
			newMsg.content_length = BitConverter.ToInt16(bytes, 5);
			if (bytes.Length > 9 + newMsg.content_length) {
				Array.Copy (bytes, 0, newMsg.full_content, 0, 9+newMsg.content_length);
			}
		}

		return newMsg;
	}

	public bool loadContent(byte[] bytes) {
		if (bytes.Length > this.content_length) {
			byte[] new_content = new byte[9 + this.content_length];
			Array.Copy (this.full_content, 0, new_content, 0, 9);
			Array.Copy(bytes, 0, new_content, 9, this.content_length);
			return true;
		}

		return false;
	}
}

public class LoginMessage : NetMessage {
	new public byte message_type = 1;

	public LoginMessage(string password, string username) {
		// TODO: need usr/pass separator
		this.content = Encoding.ASCII.GetBytes(username + password);
		this.content_length = (Int16)content.Length;
	}
}

public class ThrustSetMessage : NetMessage {
	new public byte message_type = 4;

	public ThrustSetMessage(float[] thrust_percents) {
	}
}

