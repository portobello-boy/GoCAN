package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"main/data"

	"github.com/go-chi/chi"
)

type Server struct {
	Reg  *Region
	C    *http.Client
	Port string
}

func CreateServer(dim, red int, port string) *Server {
	serv := new(Server)
	serv.Reg = CreateRegion(dim, red)
	serv.C = &http.Client{}
	serv.Port = port
	return serv
}

func (s *Server) Join(w http.ResponseWriter, r *http.Request) {
	// Add JSON headers and parse body to appropriate type
	w.Header().Add("Content-Type", "application/json")
	jr := data.ParseJoin(w, r)
	pt := HashStringToPoint(jr.Key, s.Reg.Dimension)

	// Determine if hashed point is in this region
	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Join request received, splitting region...")
		newReg, delHosts := s.Reg.Split(r.Host)

		// Encode the response to JSON body and send it
		jRes := &data.JoinResponse{
			Dimension:  newReg.Dimension,
			Redundancy: newReg.Redundancy,
			Range:      *(newReg.Space.GetRangeResponse()),
			Data:       newReg.Data,
			Neighbors:  newReg.GetNeighborResponse(),
		}
		json.NewEncoder(w).Encode(jRes)

		// Update our neighbors with our new region
		neighborReq := &data.NeighborRequest{
			Port:  s.Port,
			Range: *(s.Reg.Space.GetRangeResponse()),
		}

		body, _ := json.Marshal(neighborReq)

		// Request existing neighbors to update my range in their map
		for hst, _ := range s.Reg.Neighbors {
			req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("http://%s:%s/neighbors", hst.IP, hst.Port), bytes.NewBuffer(body))
			_, err := s.C.Do(req)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Request neighbors that are no longer adjacent to delete me
		for _, hst := range delHosts {
			req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s:%s/neighbors?port=%s", hst.IP, hst.Port, s.Port), bytes.NewBuffer(body))
			_, err := s.C.Do(req)
			if err != nil {
				log.Fatal(err)
			}
		}

	} else {
		// Forward join request to best neighbor
		log.Print("Forwarding join request to ", neighbor.IP, ":", neighbor.Port)
		body, _ := json.Marshal(jr)
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%s/join", neighbor.IP, neighbor.Port), bytes.NewBuffer(body))

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		frwdResponse, _ := ioutil.ReadAll(resp.Body)
		w.Write(frwdResponse)
	}
}

func (s *Server) SendJoin(host, port, key string) {
	// Send a join request to an existing CAN server
	log.Print("Attempting to join network at " + host)
	jr := &data.JoinRequest{
		Key: key,
	}
	body, _ := json.Marshal(jr)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/join", host), bytes.NewBuffer(body))

	resp, err := s.C.Do(req)
	log.Print("Received join response containing new region")
	if err != nil {
		log.Fatal(err)
	}

	// Handle response
	jRes := data.JoinResponse{}
	json.NewDecoder(resp.Body).Decode(&jRes)

	s.Reg.Dimension = jRes.Dimension
	s.Reg.Redundancy = jRes.Redundancy
	s.Reg.Data = jRes.Data
	s.Reg.Space.P1.Coords = jRes.Range.P1.Coords
	s.Reg.Space.P2.Coords = jRes.Range.P2.Coords
	s.Reg.Neighbors = UnpackNeighbors(jRes.Neighbors)

	// Update our neighbors with our new region
	neighborReq := &data.NeighborRequest{
		Port:  s.Port,
		Range: *(s.Reg.Space.GetRangeResponse()),
	}

	body, _ = json.Marshal(neighborReq)

	// Tell our new neighbors to add us
	for hst := range s.Reg.Neighbors {
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s:%s/neighbors", hst.IP, hst.Port), bytes.NewBuffer(body))
		_, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (s *Server) Debug(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	dRes := &data.DebugResponse{
		Dimension:  s.Reg.Dimension,
		Redundancy: s.Reg.Redundancy,
		Range:      *(s.Reg.Space.GetRangeResponse()),
		Neighbors:  s.Reg.GetNeighborResponse(),
		Data:       s.Reg.Data,
	}
	json.NewEncoder(w).Encode(dRes)
}

func (s *Server) RouteTrace(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	dr := data.ParseData(w, r)
	pt := HashStringToPoint(dr.Key, s.Reg.Dimension)

	// Determine if the key is in region, find neighbor if not
	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Processing trace request")
		log.Print("Responding with host: ", r.Host)
		tRes := &data.TraceResponse{
			Route: []string{"dest " + r.Host},
		}
		json.NewEncoder(w).Encode(tRes)
	} else { // Forward the trace request to the appropriate neighbor
		log.Print("Forwarding trace request to ", neighbor.IP, ":", neighbor.Port)
		body, _ := json.Marshal(dr)
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s:%s/trace", neighbor.IP, neighbor.Port), bytes.NewBuffer(body))

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		tr := data.ParseTrace(w, resp)
		tr.Route = append(tr.Route, "step "+r.Host)

		// log.Print(resp)
		frwdResponse, _ := json.Marshal(tr)
		w.Write(frwdResponse)
	}

	log.Print("Trace request processed")
}

func (s *Server) PutData(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	dr := data.ParseData(w, r)
	pt := HashStringToPoint(dr.Key, s.Reg.Dimension)

	// Determine if the key is in region, find neighbor if not
	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Processing add data request")
		added, err := s.Reg.AddData(pt, dr.Key, dr.Data) // Add to this region

		// Send success/failure message
		if err != nil {
			log.Print(err)
			dRes := &data.ErrorResponse{
				Message: err.Error(),
			}
			json.NewEncoder(w).Encode(dRes)
		} else if added {
			dRes := &data.DataResponse{
				Key:     dr.Key,
				Data:    dr.Data,
				Coords:  pt.Coords,
				Message: "Data successfully added",
			}
			json.NewEncoder(w).Encode(dRes)
		}
	} else { // Forward the put request to the appropriate neighbor
		log.Print("Forwarding add request to ", neighbor.IP, ":", neighbor.Port)
		body, _ := json.Marshal(dr)
		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s:%s/data", neighbor.IP, neighbor.Port), bytes.NewBuffer(body))

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		// log.Print(resp)
		frwdResponse, _ := ioutil.ReadAll(resp.Body)
		w.Write(frwdResponse)
	}

	log.Print("Add data request processed")
}

func (s *Server) PatchData(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	dr := data.ParseData(w, r)
	pt := HashStringToPoint(dr.Key, s.Reg.Dimension)

	// Determine if the key is in region, find neighbor if not
	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Processing Modify data request")
		added, err := s.Reg.ModifyData(pt, dr.Key, dr.Data) // Add to this region

		// Send success/failure message
		if err != nil {
			log.Print(err)
			dRes := &data.ErrorResponse{
				Message: err.Error(),
			}
			json.NewEncoder(w).Encode(dRes)
		} else if added {
			dRes := &data.DataResponse{
				Key:     dr.Key,
				Data:    dr.Data,
				Coords:  pt.Coords,
				Message: "Data successfully modified",
			}
			json.NewEncoder(w).Encode(dRes)
		}
	} else { // Forward the put request to the appropriate neighbor
		log.Print("Forwarding patch request to ", neighbor.IP, ":", neighbor.Port)
		body, _ := json.Marshal(dr)
		req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("http://%s:%s/data", neighbor.IP, neighbor.Port), bytes.NewBuffer(body))

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		// log.Print(resp)
		frwdResponse, _ := ioutil.ReadAll(resp.Body)
		w.Write(frwdResponse)
	}

	log.Print("Modify data request processed")
}

func (s *Server) GetData(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	key := chi.URLParam(r, "key")
	pt := HashStringToPoint(key, s.Reg.Dimension)

	// Determine if the key is in region, find neighbor if not
	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Processing data retrieval request")
		got, datum, err := s.Reg.GetData(pt, key)

		// Send success/failure message
		if err != nil {
			log.Print(err)
			dRes := &data.ErrorResponse{
				Message: err.Error(),
			}
			json.NewEncoder(w).Encode(dRes)
		} else if got {
			dRes := &data.DataResponse{
				Key:     key,
				Data:    datum,
				Coords:  pt.Coords,
				Message: "Data successfully retrieved",
			}
			json.NewEncoder(w).Encode(dRes)
		}

	} else { // Forward the get request to the appropriate neighbor
		log.Print("Forwarding get request to ", neighbor.IP, ":", neighbor.Port)
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s/data/%s", neighbor.IP, neighbor.Port, key), nil)

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		// log.Print(resp)
		frwdResponse, _ := ioutil.ReadAll(resp.Body)
		w.Write(frwdResponse)
	}

	log.Print("Get data request processed")
}

func (s *Server) DeleteData(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	key := chi.URLParam(r, "key")
	pt := HashStringToPoint(key, s.Reg.Dimension)

	// Determine if the key is in region, find neighbor if not
	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Processing data delete request")
		deleted, datum, err := s.Reg.DeleteData(pt, key)

		// Send success/failure message
		if err != nil {
			log.Print(err)
			dRes := &data.ErrorResponse{
				Message: err.Error(),
			}
			json.NewEncoder(w).Encode(dRes)
		} else if deleted {
			dRes := &data.DataResponse{
				Key:     key,
				Data:    datum,
				Coords:  pt.Coords,
				Message: "Data successfully deleted",
			}
			json.NewEncoder(w).Encode(dRes)
		}

	} else { // Forward the get request to the appropriate neighbor
		log.Print("Forwarding get request to ", neighbor.IP, ":", neighbor.Port)
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s:%s/data/%s", neighbor.IP, neighbor.Port, key), nil)

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		// log.Print(resp)
		frwdResponse, _ := ioutil.ReadAll(resp.Body)
		w.Write(frwdResponse)
	}

	log.Print("Delete data request processed")
}

func (s *Server) AddNeighbor(w http.ResponseWriter, r *http.Request) {
	nr := data.ParseNeighbor(w, r)
	nHost, _ := getHostFromRemoteAddr(r.RemoteAddr)
	err := s.Reg.AddNeighbor(nHost, nr.Port, *UnpackRange(nr.Range))
	if err != nil {
		log.Print(err)
		dRes := &data.ErrorResponse{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(dRes)
	}
}

func (s *Server) DeleteNeighbor(w http.ResponseWriter, r *http.Request) {
	nPort := r.URL.Query().Get("port")
	nHost, _ := getHostFromRemoteAddr(r.RemoteAddr)
	host := Host{
		IP:   nHost,
		Port: nPort,
	}

	_, prs := s.Reg.Neighbors[host]
	if !prs {
		err := errors.New("Host does not exist in neighbor map")
		log.Print(err)
		dRes := &data.ErrorResponse{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(dRes)
	} else {
		log.Print("Deleting neighbor from map")
		delete(s.Reg.Neighbors, host)
	}
}

func (s *Server) PatchNeighbor(w http.ResponseWriter, r *http.Request) {
	nr := data.ParseNeighbor(w, r)
	nHost, _ := getHostFromRemoteAddr(r.RemoteAddr)

	host := Host{
		IP:   nHost,
		Port: nr.Port,
	}

	_, prs := s.Reg.Neighbors[host]
	if !prs {
		err := errors.New("Host does not exist in neighbor map")
		log.Print(err)
		dRes := &data.ErrorResponse{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(dRes)
	} else {
		log.Print("Updating range for neighbor")
		s.Reg.Neighbors[host] = *UnpackRange(nr.Range)
	}
}

func (s *Server) Options(w http.ResponseWriter, r *http.Request) {}
func (s *Server) JoinOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Allow", "OPTIONS, POST")
}
func (s *Server) DataOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Allow", "OPTIONS, GET, DELETE, PUT, PATCH")
}
func (s *Server) DebugOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Allow", "OPTIONS, GET")
}

func getHostFromRemoteAddr(remoteAddr string) (string, string) {
	r := regexp.MustCompile(`^(\[::1\]):([0-9]*)$`) // Handle [::1] = localhost in IPv6
	if r.MatchString(remoteAddr) {
		splt := r.FindStringSubmatch(remoteAddr) // "[::1]:12345" -> ["[::1]:12345", "[::1]", "12345"]
		return "localhost", splt[2]              // splt[2] is the port
	}

	joinerInfo := strings.Split(remoteAddr, ":")
	return joinerInfo[0], joinerInfo[1]
}
