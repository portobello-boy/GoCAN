package server

import (
	"fmt"
	"main/data"
	"math"
	"strconv"
)

type Range struct {
	P1 Point `json:"p1"`
	P2 Point `json:"p2"`
}

func (r *Range) GetRangeResponse() *data.RangeResponse {
	rr := new(data.RangeResponse)
	rr.P1 = *new(data.PointResponse)
	rr.P1.Coords = r.P1.Coords
	rr.P2 = *new(data.PointResponse)
	rr.P2.Coords = r.P2.Coords
	return rr
}

func (r *Range) PointInRange(pt Point) bool {
	for i, val := range pt.Coords {
		if val < r.P1.Coords[i] || val >= r.P2.Coords[i] {
			return false
		}
	}
	return true
}

func (r *Range) Dimensions() *Point {
	return r.P2.Sub(r.P1)
}

func (r *Range) Split() *Range {
	splitInd := 0
	for i, val := range r.Dimensions().Coords {
		if val > r.Dimensions().Coords[splitInd] {
			splitInd = i
		}
	}
	newDimVal := (r.P1.Coords[splitInd] + r.P2.Coords[splitInd]) / 2

	newRange := new(Range)
	newRange.P1 = *(r.P1.Copy())
	newRange.P1.Coords[splitInd] = newDimVal
	newRange.P2 = *(r.P2.Copy())

	r.P2.Coords[splitInd] = newDimVal

	return newRange
}

func StringPadZero(s string, len int) string {
	return fmt.Sprintf("%0"+strconv.Itoa(len)+"s", s)
}

func (r *Range) GetCorners() []Point {
	dim := r.Dimensions()
	numDim := len(dim.Coords)
	perms := []string{}
	for i := 0; i < int(math.Pow(2, float64(numDim))); i++ {
		perms = append(perms, StringPadZero(strconv.FormatInt(int64(i), 2), numDim))
	}

	corners := []Point{}
	for _, perm := range perms {
		pt := *new(Point)
		for i, char := range perm {
			if char == '0' {
				pt.Coords = append(pt.Coords, r.P1.Coords[i])
			} else {
				pt.Coords = append(pt.Coords, r.P2.Coords[i])
			}
		}
		corners = append(corners, pt)
	}
	return corners
}

func (r *Range) PointInside(pt *Point) bool {
	for i := range r.P1.Coords {
		if !(r.P1.Coords[i] <= pt.Coords[i] && pt.Coords[i] <= r.P2.Coords[i]) {
			// log.Print(pt, " is not inside ", r)
			return false
		}
		// log.Print(pt, " is inside ", r)
	}
	return true
}

// DirectionalBorder - r is the "smaller" volume, other is the "larger" volume
func (r *Range) DirectionalBorder(other *Range) bool {
	contacts := 0
	dim := len(r.Dimensions().Coords)
	minContactPoints := int(math.Pow(2, float64(dim-1)))

	for _, point := range r.GetCorners() {
		if other.PointInside(&point) {
			contacts++
		}
		if contacts >= minContactPoints {
			return true
		}
	}
	return false
}

func (r *Range) Neighbors(other *Range) bool {
	return r.DirectionalBorder(other) || other.DirectionalBorder(r)
}

func UnpackRange(rr data.RangeResponse) *Range {
	r := new(Range)
	r.P1 = *new(Point)
	r.P1.Coords = rr.P1.Coords
	r.P2 = *new(Point)
	r.P2.Coords = rr.P2.Coords
	return r
}
