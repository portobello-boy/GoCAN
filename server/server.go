package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"main/data"

	"github.com/go-chi/chi"
)

type Server struct {
	Reg *Region
	C   *http.Client
}

func CreateServer(dim, red int) *Server {
	serv := new(Server)
	serv.Reg = CreateRegion(dim, red)
	serv.C = &http.Client{}
	return serv
}

func (s *Server) PutData(w http.ResponseWriter, r *http.Request) {
	dr := data.ParseData(w, r)
	pt := HashStringToPoint(dr.Key, s.Reg.Dimension)

	// Determine if the key is in region, find neighbor if not
	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Processing add data request")
		added, err := s.Reg.AddData(pt, dr.Key, dr.Data) // Add to this region
		w.Header().Add("Content-Type", "application/json")

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
		req, err := http.NewRequest(http.MethodPut, neighbor.IP+neighbor.Port, bytes.NewBuffer(body))

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		log.Print(resp)
	}

	log.Print("Add data request processed")
}

func (s *Server) GetData(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	pt := HashStringToPoint(key, s.Reg.Dimension)

	// Determine if the key is in region, find neighbor if not
	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Processing data retrieval request")
		got, datum, err := s.Reg.GetData(pt, key)
		w.Header().Add("Content-Type", "application/json")

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
		req, err := http.NewRequest(http.MethodGet, neighbor.IP+neighbor.Port+"/"+key, nil)

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		log.Print(resp)
	}

	log.Print("Get data request processed")
}

func (s *Server) DeleteData(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	pt := HashStringToPoint(key, s.Reg.Dimension)

	// Determine if the key is in region, find neighbor if not
	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Processing data delete request")
		deleted, datum, err := s.Reg.DeleteData(pt, key)
		w.Header().Add("Content-Type", "application/json")

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
		req, err := http.NewRequest(http.MethodDelete, neighbor.IP+neighbor.Port+"/"+key, nil)

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		log.Print(resp)
	}

	log.Print("Delete data request processed")
}

func (s *Server) PatchData(w http.ResponseWriter, r *http.Request) {
	dr := data.ParseData(w, r)
	pt := HashStringToPoint(dr.Key, s.Reg.Dimension)

	// Determine if the key is in region, find neighbor if not
	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Processing Modify data request")
		added, err := s.Reg.ModifyData(pt, dr.Key, dr.Data) // Add to this region
		w.Header().Add("Content-Type", "application/json")

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
		log.Print("Forwarding add request to ", neighbor.IP, ":", neighbor.Port)
		body, _ := json.Marshal(dr)
		req, err := http.NewRequest(http.MethodPut, neighbor.IP+neighbor.Port, bytes.NewBuffer(body))

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		log.Print(resp)
	}

	log.Print("Modify data request processed")
}

func (s *Server) Debug(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	dRes := &data.DebugResponse{
		Dimension:  s.Reg.Dimension,
		Redundancy: s.Reg.Redundancy,
		Range:      *(s.Reg.Space.GetRangeResponse()),
		Data:       s.Reg.Data,
	}
	json.NewEncoder(w).Encode(dRes)
}

func (s *Server) Join(w http.ResponseWriter, r *http.Request) {
	jr := data.ParseJoin(w, r)
	pt := HashStringToPoint(jr.Key, s.Reg.Dimension)

	inReg, neighbor := s.Reg.Locate(pt)
	if inReg {
		log.Print("Join request received, splitting region...")
		newReg := s.Reg.Split()

		addr := r.RemoteAddr
		joinerInfo := strings.Split(addr, ":")
		s.Reg.AddNeighbor(joinerInfo[0], joinerInfo[1], newReg.Space)

		// json.NewEncoder(w).Encode(newReg)
		dRes := &data.DebugResponse{
			Dimension:  newReg.Dimension,
			Redundancy: newReg.Redundancy,
			Range:      *(newReg.Space.GetRangeResponse()),
			Data:       newReg.Data,
		}
		json.NewEncoder(w).Encode(dRes)
	} else {
		log.Print("Forwarding join request to ", neighbor.IP, ":", neighbor.Port)
		body, _ := json.Marshal(jr)
		req, err := http.NewRequest(http.MethodPost, neighbor.IP+neighbor.Port, bytes.NewBuffer(body))

		resp, err := s.C.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		log.Print(resp)
	}
}
