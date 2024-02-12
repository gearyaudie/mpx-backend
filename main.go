package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gearyaudie/mpx-backend.git/db"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client // Declare a global variable to hold the MongoDB client

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get MongoDB URI from environment variable
	mongoURI := os.Getenv("MONGODB_URI") + "?tls=true"
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable not set.")
	}

	var err error
	client, err = db.GetMongoClient(mongoURI)
	if err != nil {
		log.Fatal("Error initializing MongoDB client:", err)
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
