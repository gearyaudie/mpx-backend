package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

// var client *mongo.Client

func addProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	productName := r.FormValue("name")
	description := r.FormValue("desc")

	// Create a Product struct with form values
	product := Product{
		Name: productName,
		Desc: description,
		// Add other fields as needed
	}

	// Access the file
	file, handler, err := r.FormFile("img")
	if err != nil {
		http.Error(w, "Unable to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a new GridFS file
	bucket, err := gridfs.NewBucket(
		client.Database("mpxDB"),
	)
	if err != nil {
		http.Error(w, "Unable to create GridFS bucket", http.StatusInternalServerError)
		return
	}

	// Upload the file to GridFS
	fileID, err := bucket.UploadFromStream(handler.Filename, file)
	if err != nil {
		http.Error(w, "Unable to upload file to GridFS", http.StatusInternalServerError)
		return
	}

	// Set the ImagePath field in the Product struct
	product.Img = fileID.Hex()

	// Insert the Product into the database
	var productCollection = client.Database("mpxDB").Collection("products")
	insertResult, err := productCollection.InsertOne(context.TODO(), product)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to insert product", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(insertResult.InsertedID)
}
func getAllProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var products []Product

	collection := client.Database("mpxDB").Collection("products")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var product Product
		cursor.Decode(&product)
		products = append(products, product)
	}

	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	json.NewEncoder(w).Encode(products)
}
