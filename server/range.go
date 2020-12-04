package server

import (
	"main/data"
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

func (r *Range) Neighbors(other *Range) bool {
	return true
}

func UnpackRange(rr data.RangeResponse) *Range {
	r := new(Range)
	r.P1 = *new(Point)
	r.P1.Coords = rr.P1.Coords
	r.P2 = *new(Point)
	r.P2.Coords = rr.P2.Coords
	return r
}
