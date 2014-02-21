using UnityEngine;
using System.Collections;

public class Tracker : MonoBehaviour {
	public GameObject tracked;
	public GameObject background;
	private float scale = 0.45f;

	// Use this for initialization
	void Start () {
	
	}
	
	// Update is called once per frame
	void Update () {
		if (Input.GetAxis("Mouse ScrollWheel") < 0) // back
		{
			if (Camera.main.orthographicSize < 10) {
				((Camera)this.camera).orthographicSize += 1;
				this.background.transform.localScale = new Vector3(this.background.transform.localScale.x+scale,this.background.transform.localScale.y+scale,1);
			}
		}
		if (Input.GetAxis("Mouse ScrollWheel") > 0) // forward
		{
			if (Camera.main.orthographicSize > 2) {
				((Camera)this.camera).orthographicSize -= 1;
				this.background.transform.localScale = new Vector3(this.background.transform.localScale.x-scale,this.background.transform.localScale.y-scale,1);
			}
		}
		this.transform.position = new Vector3(tracked.transform.position.x, tracked.transform.position.y, this.transform.position.z);
		this.background.transform.position = new Vector3(tracked.transform.position.x, tracked.transform.position.y, this.background.transform.position.z);
	}
}
