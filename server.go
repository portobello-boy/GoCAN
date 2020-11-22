package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"main/data"
	"main/region"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	dimFlag := flag.Int("d", 2, "Number of dimensions for this CAN server")
	redFlag := flag.Int("r", 1, "Copies of data inserted")
	flag.Parse()

	// Create region
	reg := region.CreateServer(*dimFlag, *redFlag)

	// Configure the router and client
	c := &http.Client{}
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Endpoints
	r.Post("/join", func(w http.ResponseWriter, r *http.Request) {
		jr := data.ParseJoin(w, r)
		pt := region.HashStringToPoint(jr.Key, *dimFlag)

		inReg, neighbor := reg.Locate(pt)
		if inReg {
			log.Print("Join request received, splitting region...")

		} else {
			log.Print("Forwarding join request to ", neighbor.IP, ":", neighbor.Port)
			body, _ := json.Marshal(jr)
			req, err := http.NewRequest(http.MethodPost, neighbor.IP+neighbor.Port, bytes.NewBuffer(body))

			resp, err := c.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			log.Print(resp)
		}
	})

	// Retrieve data from the CAN
	r.Get("/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		pt := region.HashStringToPoint(key, *dimFlag)

		// Determine if the key is in region, find neighbor if not
		inReg, neighbor := reg.Locate(pt)
		if inReg {
			log.Print("Processing data retrieval request")
			got, datum, err := reg.GetData(pt, key)
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
					Key:    key,
					Data:   datum,
					Coords: pt.Coords,
					Message: "Data successfully retrieved",
				}
				json.NewEncoder(w).Encode(dRes)
			}

		} else { // Forward the get request to the appropriate neighbor
			log.Print("Forwarding get request to ", neighbor.IP, ":", neighbor.Port)
			req, err := http.NewRequest(http.MethodGet, neighbor.IP+neighbor.Port+"/"+key, nil)

			resp, err := c.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			log.Print(resp)
		}

		log.Print("Get data request processed")
	})

	// Delete data from CAN
	r.Delete("/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "key")
		pt := region.HashStringToPoint(key, *dimFlag)

		// Determine if the key is in region, find neighbor if not
		inReg, neighbor := reg.Locate(pt)
		if inReg {
			log.Print("Processing data delete request")
			deleted, datum, err := reg.DeleteData(pt, key)
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
					Key:    key,
					Data:   datum,
					Coords: pt.Coords,
					Message: "Data successfully deleted",
				}
				json.NewEncoder(w).Encode(dRes)
			}

		} else { // Forward the get request to the appropriate neighbor
			log.Print("Forwarding get request to ", neighbor.IP, ":", neighbor.Port)
			req, err := http.NewRequest(http.MethodDelete, neighbor.IP+neighbor.Port+"/"+key, nil)

			resp, err := c.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			log.Print(resp)
		}

		log.Print("Get data request processed")
	})

	// Add data to CAN
	r.Put("/data", func(w http.ResponseWriter, r *http.Request) {
		dr := data.ParseData(w, r)
		pt := region.HashStringToPoint(dr.Key, *dimFlag)

		// Determine if the key is in region, find neighbor if not
		inReg, neighbor := reg.Locate(pt)
		if inReg {
			log.Print("Processing add data request")
			added, err := reg.AddData(pt, dr.Key, dr.Data) // Add to this region
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
					Key:    dr.Key,
					Data:   dr.Data,
					Coords: pt.Coords,
				}
				json.NewEncoder(w).Encode(dRes)
			}
		} else { // Forward the put request to the appropriate neighbor
			log.Print("Forwarding add request to ", neighbor.IP, ":", neighbor.Port)
			body, _ := json.Marshal(dr)
			req, err := http.NewRequest(http.MethodPut, neighbor.IP+neighbor.Port, bytes.NewBuffer(body))

			resp, err := c.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			log.Print(resp)
		}

		log.Print("Add data request processed")
	})

	r.Delete("/data", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("deleting data..."))
	})

	log.Print("Server listening on port 3000...")
	http.ListenAndServe(":3000", r)
}
