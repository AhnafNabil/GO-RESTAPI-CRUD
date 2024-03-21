package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	//"math/rand"
	"net/http"
	//"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectString = "mongodb+srv://ahnafnabil14:OzActIYwh2DRkxDu@cluster0.rkmkiqc.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
	dbName        = "Go-movies"
	colName       = "crud-operation"
)

type Movie struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Isbn     string             `json:"isbn" bson:"isbn"`
	Title    string             `json:"title" bson:"title"`
	Director *Director          `json:"director" bson:"director"`
}

type Director struct {
	Firstname string `json:"firstname" bson:"firstname"`
	Lastname  string `json:"lastname" bson:"lastname"`
}

var collection *mongo.Collection

// mongodb+srv://ahnafnabil14:<password>@cluster0.rkmkiqc.mongodb.net/
// mongodb+srv://ahnafnabil14:OzActIYwh2DRkxDu@cluster0.rkmkiqc.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0
// OzActIYwh2DRkxDu

func init() {
	clientOptions := options.Client().ApplyURI(connectString)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	collection = client.Database(dbName).Collection(colName)
}

func getMovies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch movies", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())
	var movies []Movie
	for cursor.Next(context.Background()) {
		var movie Movie
		if err := cursor.Decode(&movie); err != nil {
			http.Error(w, "Failed to decode movie", http.StatusInternalServerError)
			return
		}
		movies = append(movies, movie)
	}

	if err := cursor.Err(); err != nil {
		http.Error(w, "Cursor error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(movies); err != nil {
		http.Error(w, "Failed to get all movies", http.StatusInternalServerError)
		return
	}
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	objectID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	var movie Movie
	if err := collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&movie); err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(movie)
}

func createMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	movie.ID = primitive.NewObjectID()

	_, err := collection.InsertOne(context.Background(), movie)
	if err != nil {
		http.Error(w, "Failed to insert movie", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(movie)
}

func updateMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	movieID := params["id"]

	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectID}

	_, err = collection.ReplaceOne(context.Background(), filter, movie)
	if err != nil {
		http.Error(w, "Failed to update movie", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(movie)
}

func deleteMovie(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	movieID := params["id"]

	objectID, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objectID}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, "Failed to delete movie", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Movie deleted successfully"))
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/movies", getMovies).Methods("GET")
	r.HandleFunc("/movies/{id}", getMovie).Methods("GET")
	r.HandleFunc("/movies", createMovie).Methods("POST")
	r.HandleFunc("/movies/{id}", updateMovie).Methods("PUT")
	r.HandleFunc("/movies/{id}", deleteMovie).Methods("DELETE")

	fmt.Printf("Starting server at port 8000\n")
	log.Fatal(http.ListenAndServe(":8000", r))
}
