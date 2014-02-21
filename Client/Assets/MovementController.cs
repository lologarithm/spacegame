using UnityEngine;
using System.Collections;

public class MovementController : MonoBehaviour {
	private Vector3 axis = new Vector3(0,0,-1);
	public static float rad_convert = Mathf.PI / 180.0f;

	// Use this for initialization
	void Start () {
		Debug.Log ("test");
	}
	
	// Update is called once per frame
	void Update () {
		if (Input.GetKey(KeyCode.W)) {
			Vector3 angles = this.transform.eulerAngles;
			Vector3 vel = this.rigidbody2D.velocity;
			this.rigidbody2D.AddForce(new Vector2(Mathf.Sin(rad_convert * angles.z) * -1, Mathf.Cos (rad_convert * angles.z)) * 1);
		} 

		if (Input.GetKey(KeyCode.S)) {
			Vector3 angles = this.transform.eulerAngles;
			Vector3 vel = this.rigidbody2D.velocity;
			if (vel.magnitude > 10) {
				this.rigidbody2D.AddForce(new Vector2(Mathf.Sin(rad_convert * angles.z) * 1, Mathf.Cos (rad_convert * angles.z)) * -1);
			}
		}

		if (Input.GetKey(KeyCode.A)) {
			this.rigidbody2D.AddTorque(10f);
		}

		if (Input.GetKey(KeyCode.D)) {
			this.rigidbody2D.AddTorque(-10f);
		}
	}
}
