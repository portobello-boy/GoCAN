package server

import (
	"fmt"
	"main/data"
	"math"
	"strconv"
)

// Range - Contains two boundary points to difine a space in d-dimensions
type Range struct {
	P1 Point `json:"p1"`
	P2 Point `json:"p2"`
}

// GetRangeResponse - Marshal a range into a transmittable JSON form
func (r *Range) GetRangeResponse() *data.RangeResponse {
	rr := &data.RangeResponse{
		P1: data.PointResponse{
			Coords: r.P1.Coords,
		},
		P2: data.PointResponse{
			Coords: r.P2.Coords,
		},
	}
	return rr
}

// PointInRange - Determine if a point is inside of a space (for data interface)
func (r *Range) PointInRange(pt Point) bool {
	for i, val := range pt.Coords {
		if val < r.P1.Coords[i] || val >= r.P2.Coords[i] {
			return false
		}
	}
	return true
}

// Dimensions - Returns a normalised point containing the dimensions of a range
func (r *Range) Dimensions() *Point {
	return r.P2.Sub(r.P1)
}

// Split - Split a range into two halves, returning the new range
func (r *Range) Split() *Range {
	splitInd := 0
	for i, val := range r.Dimensions().Coords {
		if val > r.Dimensions().Coords[splitInd] {
			splitInd = i
		}
	}
	newDimVal := (r.P1.Coords[splitInd] + r.P2.Coords[splitInd]) / 2

	newRange := &Range{
		P1: *(r.P1.Copy()),
		P2: *(r.P2.Copy()),
	}

	newRange.P1.Coords[splitInd] = newDimVal
	r.P2.Coords[splitInd] = newDimVal

	return newRange
}

// stringPadZero - Pad a string s with enough zeroes to make it length len
func stringPadZero(s string, len int) string {
	return fmt.Sprintf("%0"+strconv.Itoa(len)+"s", s)
}

// GetCorners - Returns a list containing all corners defined by a range
func (r *Range) GetCorners() []Point {
	dim := r.Dimensions()
	numDim := len(dim.Coords)
	perms := []string{}
	for i := 0; i < int(math.Pow(2, float64(numDim))); i++ {
		perms = append(perms, stringPadZero(strconv.FormatInt(int64(i), 2), numDim))
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

// PointInside - Determines if a point is contained in a given range (for neighbor updates)
func (r *Range) PointInside(pt *Point) bool {
	for i := range r.P1.Coords {
		if !(r.P1.Coords[i] <= pt.Coords[i] && pt.Coords[i] <= r.P2.Coords[i]) {
			return false
		}
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

// Neighbors - Determine if two ranges share a face
func (r *Range) Neighbors(other *Range) bool {
	return r.DirectionalBorder(other) || other.DirectionalBorder(r)
}

// UnpackRange - Unmarshal a RangeResponse into a range
func UnpackRange(rr data.RangeResponse) *Range {
	r := &Range{
		P1: Point{rr.P1.Coords},
		P2: Point{rr.P2.Coords},
	}
	return r
}
