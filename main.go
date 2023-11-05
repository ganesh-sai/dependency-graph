package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
)

type Service struct {
	Name     string   `yaml:"name"`
	Version  string   `yaml:"version"`
	Kind     string   `yaml:"kind"`
	Consumes Consumes `yaml:"consumes"`
}

type Consumes struct {
	Services []ServiceConsumes `yaml:"consumes"`
}

type ServiceConsumes struct {
	Name    string   `yaml:"name"`
	Version string   `yaml:"version"`
	URL     []string `yaml:"url"`
}

const (
	DatabaseName   = "servicesdb"
	CollectionName = "services"
)

var client *mongo.Client
var servicesCollection *mongo.Collection

func init() {
	mongoDBURL := os.Getenv("MONGODB_URL")
	if mongoDBURL == "" {
		log.Fatal("MONGODB_URL environment variable is not set")
	}

	clientOptions := options.Client().ApplyURI(mongoDBURL)
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	servicesCollection = client.Database(DatabaseName).Collection(CollectionName)
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/services", createService).Methods("POST")
	r.HandleFunc("/services/{name}/{version}", getService).Methods("GET")
	r.HandleFunc("/services/{name}/{version}", updateService).Methods("PUT")
	r.HandleFunc("/services/{name}/{version}", deleteService).Methods("DELETE")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createService(w http.ResponseWriter, r *http.Request) {
	var service Service
	if err := yaml.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := servicesCollection.InsertOne(context.Background(), service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Service created"))
}

func getService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	version := vars["version"]

	var service Service
	err := servicesCollection.FindOne(context.Background(), bson.M{"name": name, "version": version}).Decode(&service)
	if err != nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	callingServicesCursor, err := servicesCollection.Find(context.Background(), bson.M{"consumes.services": bson.M{"$elemMatch": bson.M{"name": name, "version": version}}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer callingServicesCursor.Close(context.Background())

	callingServiceNames := []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}{}
	for callingServicesCursor.Next(context.Background()) {
		var callingService Service
		if err := callingServicesCursor.Decode(&callingService); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		callingServiceNames = append(callingServiceNames, struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{callingService.Name, callingService.Version})
	}

	response := struct {
		Name            string   `json:"Name"`
		Version         string   `json:"Version"`
		Kind            string   `json:"Kind"`
		Consumes        Consumes `json:"Consumes"`
		CallingServices []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"calling_services"`
	}{
		Name:            service.Name,
		Version:         service.Version,
		Kind:            service.Kind,
		Consumes:        service.Consumes,
		CallingServices: callingServiceNames,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	version := vars["version"]

	var service Service
	if err := yaml.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	filter := bson.M{"name": name, "version": version}
	update := bson.M{"$set": service}
	_, err := servicesCollection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service updated"))
}

func deleteService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	version := vars["version"]

	callingServicesCursor, err := servicesCollection.Find(context.Background(), bson.M{"consumes.services": bson.M{"$elemMatch": bson.M{"name": name, "version": version}}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	callingServiceCount := 0
	for callingServicesCursor.Next(context.Background()) {
		callingServiceCount++
	}

	if callingServiceCount > 0 {
		forceDelete := r.URL.Query().Get("forceDelete")
		if forceDelete != "true" {
			http.Error(w, "Service is being consumed by other services. Use forceDelete=true to delete it.", http.StatusBadRequest)
			return
		}
	}

	filter := bson.M{"name": name, "version": version}
	_, err = servicesCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service deleted"))
}
