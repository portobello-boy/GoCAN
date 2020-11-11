package main

import (
	"flag"
	"fmt"
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
	fmt.Println(reg.Space)

	// Configure the router
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Endpoints
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Post("/join", func(w http.ResponseWriter, r *http.Request) {
		jr := data.ParseJoin(w, r)
		pt := region.HashStringToPoint(jr.Key, *dimFlag)
		if reg.Locate(pt) {
			fmt.Println("Point " + pt + " in region " + reg.Space)
		}
		// fmt.Println(jr)
		// fmt.Println(region.HashString(jr.Key, *dimFlag))
		// reg.HandleJoinRequest(jr.Key, jr.Host)
	})

	r.Put("/data", func(w http.ResponseWriter, r *http.Request) {
		dr := data.ParseData(w, r)
		pt := region.HashStringToPoint(dr.Key, *dimFlag)
		fmt.Println(dr)
	})

	r.Delete("/data", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("deleting data..."))
	})

	http.ListenAndServe(":3000", r)
}
