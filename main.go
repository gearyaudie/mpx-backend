package main

import (
	"log"
	"net/http"

	"github.com/gearyaudie/mpx-backend.git/db"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client // Declare a global variable to hold the MongoDB client

func init() {
	// Initialize the MongoDB client during the application startup
	var err error
	client, err = db.GetMongoClient()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	route := mux.NewRouter()

	s := route.PathPrefix("/api").Subrouter() // Base Path

	// Routes
	s.HandleFunc("/addProduct", addProduct).Methods("POST")
	s.HandleFunc("/getAllProducts", getAllProducts).Methods("GET")

	// Close the MongoDB client when the application exits
	defer client.Disconnect(nil)

	log.Fatal(http.ListenAndServe(":8000", route)) // Run Server
}
