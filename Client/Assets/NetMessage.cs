//------------------------------------------------------------------------------
// <auto-generated>
//     This code was generated by a tool.
//     Runtime Version:4.0.30319.18052
//
//     Changes to this file may cause incorrect behavior and will be lost if
//     the code is regenerated.
// </auto-generated>
//------------------------------------------------------------------------------
using System;
using System.IO;

public class NetMessage {
	public byte message_type;
	public Int32 from_player;
	public Int32 content_length;
}

public class LoginMessage : NetMessage {
	string password;
	string username;

	public LoginMessage(string password, string username) {
		this.message_type = 1;
		this.username = username;
		this.password = password;
	}

	public byte[] MessageBytes() {
		///byte[] byte_array = new byte[]
		MemoryStream stream = new MemoryStream();
		using (BinaryWriter writer = new BinaryWriter(stream))
		{
			writer.Write(this.message_type);
			writer.Write(0);
			writer.Write(this.password.Length);
			writer.Write(this.password);
		}
		return stream.ToArray();
	}
}

public class ThrustSetMessage {
}
