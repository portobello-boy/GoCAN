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
	serv := server.CreateServer(*dimFlag, *redFlag)
	if *join != "" {
		fmt.Print("What key to use to join server? ")
		text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		serv.SendJoin(*join, text)
		log.Print(serv.Reg)
	}

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

	r.Options("/*", serv.Options)
	r.Options("/join", serv.JoinOptions)
	r.Options("/data", serv.DataOptions)
	r.Options("/debug", serv.DebugOptions)

	log.Print("Server listening on port " + *port + "...")
	http.ListenAndServe(":"+*port, r)
}
