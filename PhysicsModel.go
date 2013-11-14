package main

func CrossProduct(a *Vect2, b *Vect2) float32 {
	return a.X()*b.Y() - a.Y()*b.X()
}

func CrossScalar(v *Vect2, s float32) *Vect2 {
	return &Vect2{v.Y() * s, -s * v.X()}
}

func CrossScalarFirst(s float32, v *Vect2) *Vect2 {
	return &Vect2{v.Y() * -s, s * v.X()}
}

func MultVect2(a *Vect2, s float32) *Vect2 {
	return &Vect2{a.X() * s, a.Y() * s}
}

type Vect2 []float32

// Modifies this vector
func (v *Vect2) Add(v2 *Vect2) {
	[]float32(*v)[0] += []float32(*v2)[0]
	[]float32(*v2)[1] += []float32(*v2)[1]
}

func (v *Vect2) X() float32 {
	return []float32(*v)[0]
}

func (v *Vect2) Y() float32 {
	return []float32(*v)[1]
}
