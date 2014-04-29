using System;
using System.IO;
using System.Text;


public class NetMessage {
	public const int DEFAULT_FRAME_LEN = 5;
	public const byte ECHO = 0;
	public const byte LOGINREQUEST = 1;
	public const byte LOGINSUCCESS = 2;
	public const byte LOGINFAIL = 3;
	public const byte PHYSICS = 4;
	public const byte DISCONNECT = 255;

	public byte message_type;
	public Int32 from_player;
	public UInt16 content_length;
	public UInt16 sequence;
	public byte[] content;
	public byte[] full_content;


	public byte[] MessageBytes() {
		///byte[] byte_array = new byte[]
		MemoryStream stream = new MemoryStream();
		using (BinaryWriter writer = new BinaryWriter(stream))
		{
			writer.Write(this.message_type);
			writer.Write(sequence);
			writer.Write(content_length);
			writer.Write(content);
		}
		return stream.ToArray();
	}

	public byte[] Content() {
		byte[] content = null;
		Array.Copy (this.full_content, DEFAULT_FRAME_LEN, content, 0, this.full_content.Length - DEFAULT_FRAME_LEN);
		return content;
	}

	public static NetMessage fromBytes(byte[] bytes) {
		NetMessage newMsg = null;
		if (bytes.Length >= DEFAULT_FRAME_LEN) {
			newMsg = new NetMessage ();
			newMsg.message_type = bytes[0];
			newMsg.sequence = BitConverter.ToUInt16(bytes, 1);
			newMsg.content_length = BitConverter.ToUInt16(bytes, 5);
			if (bytes.Length > DEFAULT_FRAME_LEN + newMsg.content_length) {
				Array.Copy (bytes, 0, newMsg.full_content, 0, DEFAULT_FRAME_LEN+newMsg.content_length);
			}
		}

		return newMsg;
	}

	public bool loadContent(byte[] bytes) {
		if (bytes.Length >= this.content_length) {
			byte[] new_content = new byte[DEFAULT_FRAME_LEN + this.content_length];
			Array.Copy (this.full_content, 0, new_content, 0, DEFAULT_FRAME_LEN);
			Array.Copy(bytes, 0, new_content, DEFAULT_FRAME_LEN, this.content_length);
			return true;
		}

		return false;
	}
}

public class LoginMessage : NetMessage {
	public LoginMessage(string password, string username) {
		this.message_type = 1;
		this.content = Encoding.ASCII.GetBytes(username + ":" + password);
		this.content_length = (UInt16)content.Length;
	}
}

public class ThrustSetMessage : NetMessage {
	public ThrustSetMessage(float[] thrust_percents) {
		this.message_type = 5;
	}
}

public class PhysicsMessage : NetMessage {
	public PhysicsMessage() {
	}
}
