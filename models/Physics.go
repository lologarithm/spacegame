package models

func CrossProduct(a Vect2, b Vect2) float32 {
	return a.X*b.Y - a.Y*b.X
}

func CrossScalar(v Vect2, s float32) Vect2 {
	return Vect2{v.Y * s, -s * v.X}
}

func CrossScalarFirst(s float32, v Vect2) Vect2 {
	return Vect2{v.Y * -s, s * v.X}
}

func MultVect2(a Vect2, s float32) Vect2 {
	return Vect2{a.X * s, a.Y * s}
}

type Vect2 struct {
	X, Y float32
}

func (v Vect2) Add(v2 Vect2) Vect2 {
	return Vect2{v.X + v2.X, v.Y + v2.Y}
}

type RigidBody struct {
	Position Vect2 // coords x,y of entity  (meters)
	Velocity Vect2 // speed in vector format (m/s)
	Force    Vect2 // Force to apply each tick.

	Angle           float32 // Current heading (radians)
	AngularVelocity float32 // speed of rotation around the Z axis (radians/sec)
	Torque          float32 // Torque to apply each tick

	Mass       float32 // mass of the ship, (kg)
	InvMass    float32 // Inverted mass for physics calcs
	Inertia    float32 // Inertia of the ship
	InvInertia float32 // Inverted Inertia for physics calcs
}
