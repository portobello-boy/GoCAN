package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"main/server"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	dimFlag := flag.Int("d", 2, "Number of dimensions for this CAN server")
	redFlag := flag.Int("r", 1, "Copies of data inserted")
	port := flag.String("p", "3000", "Port to listen on")
	join := flag.String("join", "", "IP:Port of existing server to join")

	flag.Parse()

	// Create region
	serv := server.CreateServer(*dimFlag, *redFlag, *port)
	if *join != "" {
		fmt.Print("What key to use to join server? ")
		text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		serv.SendJoin(*join, *port, text)
		// log.Print(serv.Reg)
	}

	// Configure the router and client
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)

	// Endpoints
	r.Route("/", func(r chi.Router) {
		// Join a CAN
		r.Post("/join", serv.Join)

		// Get info from CAN Server
		r.Get("/debug", serv.Debug)
		r.Post("/trace", serv.RouteTrace)

		// Interface with CAN Data
		r.Route("/data", func(r chi.Router) {
			r.Put("/", serv.PutData)            // Add data
			r.Patch("/", serv.PatchData)        // Update Data
			r.Get("/{key}", serv.GetData)       // Retrieve Data
			r.Delete("/{key}", serv.DeleteData) // Delete Data
		})

		// Interface with CAN Neighbors
		r.Route("/neighbors", func(r chi.Router) {
			r.Put("/", serv.AddNeighbor)       // Add Neighbor
			r.Patch("/", serv.PatchNeighbor)   // Update Neighbor
			r.Delete("/", serv.DeleteNeighbor) // Delete Neighbor
		})
	})

	r.Options("/*", serv.Options)
	r.Options("/join", serv.JoinOptions)
	r.Options("/data", serv.DataOptions)
	r.Options("/debug", serv.DebugOptions)

	log.Print("Server listening on port " + *port + "...")
	http.ListenAndServe(":"+*port, r)
}
