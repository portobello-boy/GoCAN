package server

import (
	"errors"
	"log"
	"main/data"
	"strings"
)

type Host struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

type Region struct {
	Dimension  int               `json:"dimension"`
	Redundancy int               `json:"redundancy"`
	Space      Range             `json:"range"`
	Data       map[string]string `json:"data"`
	Neighbors  map[Host]Range    `json:"neighbords"`
}

func CreateRegion(dim, red int) *Region {
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

func UnpackNeighbors(neighMap map[string]data.RangeResponse) map[Host]Range {
	hostMap := make(map[Host]Range)
	for hst, rng := range neighMap {
		hostInfo := strings.Split(hst, ":")
		host := new(Host)
		host.IP = hostInfo[0]
		host.Port = hostInfo[1]
		hostMap[*host] = *UnpackRange(rng)
	}
	return hostMap
}

func (r *Region) DeleteData(pt Point, key string) (bool, string, error) {
	if !r.Space.PointInRange(pt) {
		return false, "", errors.New("Point not in range")
	}

	datum, prs := r.Data[key]
	delete(r.Data, key)
	log.Print("Key:", key, ", Found:", prs)
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
	log.Print("Key:", key, ", Data:", datum)
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

func (r *Region) ModifyData(pt Point, key, val string) (bool, error) {
	if !r.Space.PointInRange(pt) {
		return false, errors.New("Point not in range")
	}

	_, prs := r.Data[key]
	if !prs {
		return false, errors.New("Key not found in map, cannot modify data")
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

func (r *Region) GetNeighborResponse() map[string]data.RangeResponse {
	nr := make(map[string]data.RangeResponse)
	for host, rng := range r.Neighbors {
		nr[host.IP+":"+host.Port] = *(rng.GetRangeResponse())
	}
	return nr
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
		dist := pt.Dist(*ran.P1.Midpoint(ran.P2))
		if dist < bestDist {
			bestDist = dist
			bestHost = &host
		}
	}

	return bestHost
}

func (r *Region) AddNeighbor(hostname, port string, rng Range) error {
	host := Host{
		IP:   hostname,
		Port: port,
	}

	_, prs := r.Neighbors[host]
	if prs {
		return errors.New("Neighbor already exists in map")
	}

	log.Print("Neighbor added to map")
	r.Neighbors[host] = rng
	return nil
}

func (r *Region) Split(myHost string) (*Region, []Host) {
	newRange := r.Space.Split()

	newReg := new(Region)
	newReg.Dimension = r.Dimension
	newReg.Redundancy = r.Redundancy
	newReg.Space = *newRange
	newReg.Data = make(map[string]string)
	newReg.Neighbors = make(map[Host]Range)

	for key, val := range r.Data {
		if pt := HashStringToPoint(key, r.Dimension); newRange.PointInRange(pt) {
			newReg.Data[key] = val
			delete(r.Data, key)
		}
	}

	delHosts := make([]Host, 0)

	myHostSplit := strings.Split(myHost, ":")
	newReg.AddNeighbor(myHostSplit[0], myHostSplit[1], r.Space)

	for host, rng := range r.Neighbors {
		if newRange.Neighbors(&rng) {
			newReg.Neighbors[host] = rng
		}
		if !r.Space.Neighbors(&rng) {
			delHosts = append(delHosts, host)
			delete(r.Neighbors, host)
		}
	}

	return newReg, delHosts
}
