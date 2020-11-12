package main

import (
	"flag"
	"log"
	"bytes"
	"encoding/json"
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
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Post("/join", func(w http.ResponseWriter, r *http.Request) {
		jr := data.ParseJoin(w, r)
		pt := region.HashStringToPoint(jr.Key, *dimFlag)

		inReg, neighbor := reg.Locate(pt)
		if inReg {
			log.Print("Join request received, splitting region...")
			
		} else {
			log.Print("Forwarding join request to ", neighbor.Ip, ":", neighbor.Port)
			body, _ := json.Marshal(jr)
			req, err := http.NewRequest(http.MethodPost, neighbor.Ip + neighbor.Port, bytes.NewBuffer(body))

			resp, err := c.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			log.Print(resp)
		}
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
					Key: dr.Key,
					Data: dr.Data,
					Coords: pt.Coords,
				}
				json.NewEncoder(w).Encode(dRes)
			}
		} else { // Forward the request to the appropriate neighbor
			log.Print("Forwarding add request to ", neighbor.Ip, ":", neighbor.Port)
			body, _ := json.Marshal(dr)
			req, err := http.NewRequest(http.MethodPut, neighbor.Ip + neighbor.Port, bytes.NewBuffer(body))

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
