package main

import (
	"flag"
	"log"
	"main/server"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	dimFlag := flag.Int("d", 2, "Number of dimensions for this CAN server")
	redFlag := flag.Int("r", 1, "Copies of data inserted")
	flag.Parse()

	// Create region
	serv := server.CreateServer(*dimFlag, *redFlag)

	// Configure the router and client
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Endpoints
	r.Post("/join", serv.Join)

	// Retrieve data from the CAN
	r.Get("/data/{key}", serv.GetData)

	// Delete data from CAN
	r.Delete("/data/{key}", serv.DeleteData)

	// Add data to CAN
	r.Put("/data", serv.PutData)

	// Update data in CAN
	r.Patch("/data", serv.PatchData)

	r.Get("/debug", serv.Debug)

	log.Print("Server listening on port 3000...")
	http.ListenAndServe(":3000", r)
}
