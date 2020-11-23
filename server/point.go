package server

import (
	"hash/fnv"
	"math"
	"strconv"
)

type Point struct {
	Coords []float64
}

func (p *Point) Copy() *Point {
	newP := new(Point)
	newP.Coords = make([]float64, len(p.Coords))
	copy(newP.Coords, p.Coords)
	return newP
}

func hash(s string) float64 {
	h := fnv.New64()
	h.Write([]byte(s))
	return float64(float64(h.Sum64()) / math.MaxUint64)
}

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

func sub(a, b Point) *Point {
	p := new(Point)
	array := make([]float64, len(a.Coords))

	for i, val := range a.Coords {
		array[i] = val - b.Coords[i]
	}
	p.Coords = array

	return p
}

func add(a, b Point) *Point {
	p := new(Point)
	array := make([]float64, len(a.Coords))

	for i, val := range a.Coords {
		array[i] = val + b.Coords[i]
	}
	p.Coords = array

	return p
}

func magnitude(pt *Point) float64 {
	sum := 0.0
	for _, val := range pt.Coords {
		sum += val * val
	}
	return math.Sqrt(sum)
}

func scale(pt *Point, scalar float64) *Point {
	p := new(Point)
	array := make([]float64, len(pt.Coords))

	for i, val := range pt.Coords {
		array[i] = val * scalar
	}
	p.Coords = array

	return p
}

func Dist(a, b Point) float64 {
	return magnitude(sub(a, b))
}

func Midpoint(a, b Point) *Point {
	return scale(add(a, b), 0.5)
}
