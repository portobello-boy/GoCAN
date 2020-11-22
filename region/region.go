package region

import (
	"errors"
	"log"
)

type Host struct {
	IP   string
	Port string
}

type Region struct {
	Dimension  int
	Redundancy int
	Space      Range
	Data       map[string]string
	Neighbors  map[Host]Range
}

func CreateServer(dim, red int) *Region {
	// Create bounding points
	p1 := new(Point)
	p2 := new(Point)
	p1.Coords = make([]float64, dim)
	p2.Coords = make([]float64, dim)

	for i := range p2.Coords {
		p2.Coords[i] = 1
	}

	// Create range using points
	r := new(Range)
	r.P1 = *p1
	r.P2 = *p2

	// Create the region
	region := new(Region)
	region.Dimension = dim
	region.Redundancy = red
	region.Space = *r
	region.Data = make(map[string]string)
	region.Neighbors = make(map[Host]Range)

	return region
}

func (r *Region) DeleteData(pt Point, key string) (bool, string, error) {
	if !r.Space.PointInRange(pt) {
		return false, "", errors.New("Point not in range")
	}

	datum, prs := r.Data[key]
	delete(r.Data, key)
	log.Print("Key:", key, "Found:", prs)
	if prs {
		return true, datum, nil
	}

	return false, "", errors.New("Key does not exist in map")
}

func (r *Region) GetData(pt Point, key string) (bool, string, error) {
	if !r.Space.PointInRange(pt) {
		return false, "", errors.New("Point not in range")
	}

	datum, prs := r.Data[key]
	log.Print("Key:", key, "Found:", prs)
	if prs {
		return true, datum, nil
	}

	return false, "", errors.New("Key does not exist in map")
}

func (r *Region) AddData(pt Point, key, val string) (bool, error) {
	if !r.Space.PointInRange(pt) {
		return false, errors.New("Point not in range")
	}

	_, prs := r.Data[key]
	if prs {
		return false, errors.New("Key already exists in map")
	}

	r.Data[key] = val
	return true, nil
}

func (r *Region) Locate(pt Point) (bool, *Host) {
	for i, val := range pt.Coords {
		// If any of the point's dimensions are outside our bounds, find a neighbor
		if val < r.Space.P1.Coords[i] || val > r.Space.P2.Coords[i] {
			host := r.findNearestNeighbor(pt)
			return false, host
		}
	}

	// Return true since the point is in our bounds
	return true, nil
}

func (r *Region) findNearestNeighbor(pt Point) *Host {
	bestDist := 1.0
	bestHost := new(Host)

	for host, ran := range r.Neighbors {
		// If the point is in a known neighbor, return them
		if ran.PointInRange(pt) {
			return &host
		}

		// Determine which neighbor's midpoint is closest to the point
		dist := Dist(pt, *Midpoint(ran.P1, ran.P2))
		if dist < bestDist {
			bestDist = dist
			bestHost = &host
		}
	}

	return bestHost
}

// func (r *Region) Split(w http.ResponseWriter, r *http.Request) {

// }
