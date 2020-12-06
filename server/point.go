package server

import (
	"hash/fnv"
	"math"
	"strconv"
)

// Point - Contains an array of d coordinates for a point in d-dimensional space
type Point struct {
	Coords []float64
}

// Copy - Duplicate a point
func (pt *Point) Copy() *Point {
	newP := &Point{
		Coords: make([]float64, len(pt.Coords)),
	}
	copy(newP.Coords, pt.Coords)
	return newP
}

// hash - Take a string and hash it to a float64 value
func hash(s string) float64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return float64(float64(h.Sum64()) / math.MaxUint64)
}

// HashStringToPoint - Hash a string into a d-dimensional point
func HashStringToPoint(key string, dim int) Point {
	array := make([]float64, dim)
	point := new(Point)

	// Hash key, and use hash for the next hash
	for i := 0; i < dim; i++ {
		hsh := hash(key)
		array[i] = hsh
		key = strconv.FormatFloat(hsh, 'f', -1, 64)
	}
	point.Coords = array

	return *point
}

// Sub - Subtract point a from point b, return a new point
func (pt *Point) Sub(b Point) *Point {
	p := new(Point)
	array := make([]float64, len(pt.Coords))

	for i, val := range pt.Coords {
		array[i] = val - b.Coords[i]
	}
	p.Coords = array

	return p
}

// Add - Add point a to point b, return a new point
func (pt *Point) Add(b Point) *Point {
	p := new(Point)
	array := make([]float64, len(pt.Coords))

	for i, val := range pt.Coords {
		array[i] = val + b.Coords[i]
	}
	p.Coords = array

	return p
}

// Magnitude - Return the magnitude of a point relative to the origin
func (pt *Point) Magnitude() float64 {
	sum := 0.0
	for _, val := range pt.Coords {
		sum += val * val
	}
	return math.Sqrt(sum)
}

// Scale - Scale a point by a given scalar, return a new point
func (pt *Point) Scale(scalar float64) *Point {
	p := new(Point)
	array := make([]float64, len(pt.Coords))

	for i, val := range pt.Coords {
		array[i] = val * scalar
	}
	p.Coords = array

	return p
}

// Dist - Return the distance between two points
func (pt *Point) Dist(b Point) float64 {
	return pt.Sub(b).Magnitude()
}

// Midpoint - Find the midpoint between two points
func (pt *Point) Midpoint(b Point) *Point {
	return pt.Add(b).Scale(0.5)
}
