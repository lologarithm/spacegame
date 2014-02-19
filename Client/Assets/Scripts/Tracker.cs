using UnityEngine;
using System.Collections;

public class Tracker : MonoBehaviour {
	public GameObject tracked;
	public GameObject background;

	// Use this for initialization
	void Start () {
	
	}
	
	// Update is called once per frame
	void Update () {
		if (Input.GetAxis("Mouse ScrollWheel") < 0) // back
		{
			((Camera)this.camera).orthographicSize = Mathf.Min(Camera.main.orthographicSize+1, 10);
		}
		if (Input.GetAxis("Mouse ScrollWheel") > 0) // forward
		{
			((Camera)this.camera).orthographicSize = Mathf.Max(Camera.main.orthographicSize-1, 2);
		}
		this.transform.position = new Vector3(tracked.transform.position.x, tracked.transform.position.y, this.transform.position.z);

	}
}
