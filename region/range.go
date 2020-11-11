package region

type Range struct {
	P1 Point
	P2 Point
}

func (r *Range) PointInRange(pt Point) bool {
	for i, val := range pt.Coords {
		if val < r.P1.Coords[i] || val >= r.P2.Coords[i] {
			return false
		}
	}
	return true
}
