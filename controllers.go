package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	w.Header().Set("Content-Type", "multipart/form-data")

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

	ctx, _ := context.WithTimeout(context.Background(), time.Second*30)
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

		// Fetch file content from GridFS using the file ID
		fileID, err := primitive.ObjectIDFromHex(product.Img)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "Invalid file ID format" }`))
			return
		}

		// Create a new GridFS bucket
		bucket, err := gridfs.NewBucket(
			client.Database("mpxDB"),
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "Unable to create GridFS bucket" }`))
			return
		}

		// Open a stream to read the file content
		fileStream, err := bucket.OpenDownloadStream(fileID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "Unable to open download stream for file" }`))
			return
		}
		defer fileStream.Close()

		// Read the file content
		fileContent, err := ioutil.ReadAll(fileStream)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "Unable to read file content" }`))
			return
		}

		// Add the file content to the product struct
		product.ImgContent = fileContent

		// Append the modified product to the products slice
		products = append(products, product)
	}

	if err := cursor.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}

	json.NewEncoder(w).Encode(products)
}

func editProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Use mux.Vars to get the URL parameters
	vars := mux.Vars(r)
	productId := vars["id"]

	// parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10mb limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	productName := r.FormValue("name")
	description := r.FormValue("desc")

	// Access the file if submitted
	file, handler, err := r.FormFile("img")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Unable to get file", http.StatusBadRequest)
		return
	}

	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	bucket, err := gridfs.NewBucket(
		client.Database("mpxDB"),
	)

	if err != nil {
		http.Error(w, "Unable to create GridFS bucket", http.StatusInternalServerError)
		return
	}

	// Create a new GridFS bucket
	var fileID primitive.ObjectID
	if file != nil {
		fileID, err = bucket.UploadFromStream(handler.Filename, file)
		if err != nil {
			http.Error(w, "Unable to upload file to GridFS", http.StatusInternalServerError)
			return
		}
	}

	// Convert productID to ObjectId
	objID, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		http.Error(w, "Invalid product ID format", http.StatusBadRequest)
		return
	}

	var productCollection = client.Database("mpxDB").Collection("products")
	filter := bson.M{"_id": objID}

	// Prepare update based on whether an image was submitted
	var update bson.M
	if file != nil {
		update = bson.M{"$set": bson.M{
			"name": productName,
			"desc": description,
			"img":  fileID.Hex(),
		}}
	} else {
		update = bson.M{"$set": bson.M{
			"name": productName,
			"desc": description,
		}}
	}

	_, err = productCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to update product", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"message": "Product updated successfully"}`))
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get product ID from URL parameters
	vars := mux.Vars(r)
	productID := vars["id"]

	// Convert productID to ObjectID
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		http.Error(w, "Invalid product ID format", http.StatusBadRequest)
		return
	}

	// Delete the product from the database
	var productCollection = client.Database("mpxDB").Collection("products")
	filter := bson.M{"_id": objID}

	_, err = productCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Unable to delete product", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{ "message": "Product deleted successfully" }`))

}
