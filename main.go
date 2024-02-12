package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

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
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable not set.")
	}

	route := mux.NewRouter()

	s := route.PathPrefix("/api").Subrouter()

	// Routes
	s.HandleFunc("/addProduct", addProduct).Methods("POST")
	s.HandleFunc("/getAllProducts", getAllProducts).Methods("GET")

	server := &http.Server{
		Addr:    ":" + port,
		Handler: route,
	}

	// Close the MongoDB client when the application exits
	defer client.Disconnect(nil)

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	// Wait for interrupt signal to gracefully shut down the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Println("Shutting down...")
	err := server.Shutdown(context.Background())
	if err != nil {
		log.Fatal("Error during server shutdown:", err)
	}
}
