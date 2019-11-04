package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

const DATABASE = "banco"
const COLLECTION = "conta"

type Conta struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	NUMERO       int                `json:"numero,omitempty bson:"numero,omitempty"`
	SALDO        int                `json:"saldo,omitempty bson:"saldo,omitempty"`
	DATAABERTURA string             `json:"dataAbertura,omitempty bson:"dataAbertura,omitempty"`
	STATUS       bool               `json:"status,omitempty bson:"status,omitempty"`
}

func createConta(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var conta Conta
	_ = json.NewDecoder(request.Body).Decode(&conta)
	collection := client.Database(DATABASE).Collection(COLLECTION)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, _ := collection.InsertOne(ctx, conta)
	json.NewEncoder(response).Encode(result)
}

func readConta(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var people []Conta
	collection := client.Database(DATABASE).Collection(COLLECTION)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	contaID := mux.Vars(request)["id"]
	if len(contaID) == 0 {
		retriveConta(ctx, collection, response, request)
	} else {
		retriveOneConta(contaID, response, request)
	}

	json.NewEncoder(response).Encode(people)
}

func retriveOneConta(contaID string, response http.ResponseWriter, request *http.Request) {

	id, _ := primitive.ObjectIDFromHex(contaID)
	var conta Conta
	collection := client.Database(DATABASE).Collection(COLLECTION)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, Conta{ID: id}).Decode(&conta)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
}

func retriveConta(ctx context.Context, collection *mongo.Collection,
	response http.ResponseWriter, request *http.Request) {
	var people []Conta
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var conta Conta
		cursor.Decode(&conta)
		people = append(people, conta)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
}
func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Seja bem vindo!")
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/conta", createConta).Methods("POST") //ok
	router.HandleFunc("/person/{id}", readConta).Methods("GET")
	http.ListenAndServe("localhost:8081", router)
}
