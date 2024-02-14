package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gearyaudie/mpx-backend.git/db"
	"github.com/gearyaudie/mpx-backend.git/models"
	"github.com/gearyaudie/mpx-backend.git/utils"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var users = make(map[string]models.User)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var inputUser models.User
	err := json.NewDecoder(r.Body).Decode(&inputUser)

	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get MongoDB URI from environment variable
	mongoURI := os.Getenv("MONGODB_URI") + "?tls=true"
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable not set.")
	}

	// Connect to MongoDB using the MONGODB_URI environment variable
	client, err := db.GetMongoClient(mongoURI)
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(nil)

	// Query the user from the database
	userCollection := client.Database("mpxDB").Collection("users")
	var storedUser models.User
	err = userCollection.FindOne(r.Context(), bson.M{"email": inputUser.Email}).Decode(&storedUser)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Compare hashed passwords
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(inputUser.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate session token and set it in a cookie
	token, err := utils.GenerateToken()
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Set the token in the response
	responseData := map[string]string{"token": token}
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		http.Error(w, "Error encoding response data", http.StatusInternalServerError)
		return
	}

	// Set the content type header
	w.Header().Set("Content-Type", "application/json")

	// Write the response with the token
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)

}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear session token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logout successful"))
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	var newUser models.User
	err := json.NewDecoder(r.Body).Decode(&newUser)

	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Hash the user's password before storing it in the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	newUser.Password = string(hashedPassword)

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get MongoDB URI from environment variable
	mongoURI := os.Getenv("MONGODB_URI") + "?tls=true"
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable not set.")
	}

	client, err := db.GetMongoClient(mongoURI)

	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return
	}
	defer client.Disconnect(nil)

	// Insert the new user into the database
	userCollection := client.Database("mpxDB").Collection("users")
	_, err = userCollection.InsertOne(r.Context(), newUser)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created successfully"))
}
