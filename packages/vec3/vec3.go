package vec3

import "math"

//Vector3 is vector in coordinate form
type Vector3 struct {
	X, Y, Z float32
}

//Plus is a = a + b ie. a.Plus(b)
//use with caution
func (a *Vector3) Plus(b Vector3) {
	*a = Vector3{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

//Difference is <vector3> = a - b ie. Difference(a,b)
func Difference(a, b Vector3) Vector3 {
	return Vector3{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

//Mult to multiply a by b ie. Vector3.Mult(magnitude)
func (a Vector3) Mult(b float32) Vector3 {
	return Vector3{a.X * b, a.Y * b, a.Z * b}
}

//Length is for finding length of vector
func (a Vector3) Length() float32 {
	return float32(math.Sqrt(float64(a.X*a.X + a.Y*a.Y + a.Z*a.Z)))
}

//SqrMag is much faster when only camparing 2 vectors length
func (a Vector3) SqrMag() float32 {
	return float32(a.X*a.X + a.Y*a.Y + a.Z*a.Z)
}

//Distance between a<Vector3> and b<Vector3>
func Distance(a, b Vector3) float32 {
	return (Difference(a, b).Length())
}

//SqrDist is square of distance between a and b
func SqrDist(a, b Vector3) float32 {
	return (Difference(a, b).SqrMag())
}

//Normalised returns Vector3
func (a Vector3) Normalised() Vector3 {
	l := a.Length()
	return Vector3{a.X / l, a.Y / l, a.Z / l}
}

//Normalise normalises & changes provided Vector3
//use with caution
func (a *Vector3) Normalise() {
	l := a.Length()
	*a = Vector3{a.X / l, a.Y / l, a.Z / l}
}
