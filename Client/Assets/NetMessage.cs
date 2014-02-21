using System;
using System.IO;
using System.Text;

public class NetMessage {
	public byte message_type;
	public Int32 from_player;
	public Int16 content_length;
	public byte[] content;

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
}

public class LoginMessage : NetMessage {

	public LoginMessage(string password, string username) {
		this.message_type = 1;
		// TODO: need usr/pass separator
		this.content = Encoding.ASCII.GetBytes(username + password);
		this.content_length = (Int16)content.Length;
	}
}

public class ThrustSetMessage {
}

