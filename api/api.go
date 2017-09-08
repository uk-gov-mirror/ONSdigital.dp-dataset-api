package api

import (
	"github.com/ONSdigital/dp-dataset-api/auth"
	"github.com/ONSdigital/dp-dataset-api/dimension"
	"github.com/ONSdigital/dp-dataset-api/instance"
	"github.com/ONSdigital/dp-dataset-api/store"
	"github.com/gorilla/mux"
)

//go:generate moq -out apitest/api.go -pkg apitest . API

//API provides an interface for the routes
type API interface {
	CreateDatasetAPI(string, *mux.Router, store.DataStore) *DatasetAPI
}

// DatasetAPI manages importing filters against a dataset
type DatasetAPI struct {
	dataStore   store.DataStore
	router      *mux.Router
	privateAuth *auth.Authenticator
	instanceStore instance.Store
}

// CreateDatasetAPI manages all the routes configured to API
func CreateDatasetAPI(secretKey string, router *mux.Router, dataStore store.DataStore) *DatasetAPI {
	router.Path("/healthcheck").Methods("GET").HandlerFunc(healthCheck)

	api := DatasetAPI{privateAuth: &auth.Authenticator{SecretKey: secretKey, HeaderName: "internal-token"}, dataStore: dataStore, router: router}
	api.router.HandleFunc("/datasets", api.getDatasets).Methods("GET")
	api.router.HandleFunc("/datasets", api.addDataset).Methods("POST")
	api.router.HandleFunc("/datasets/{id}", api.getDataset).Methods("GET")
	api.router.HandleFunc("/datasets/{id}/editions", api.getEditions).Methods("GET")
	api.router.HandleFunc("/datasets/{id}/editions/{edition}", api.getEdition).Methods("GET")
	api.router.HandleFunc("/datasets/{id}/editions/{edition}", api.addEdition).Methods("POST")
	api.router.HandleFunc("/datasets/{id}/editions/{edition}/versions", api.getVersions).Methods("GET")
	api.router.HandleFunc("/datasets/{id}/editions/{edition}/versions/{version}", api.getVersion).Methods("GET")
	api.router.HandleFunc("/datasets/{id}/editions/{edition}/versions/{version}", api.addVersion).Methods("POST")

	instance := instance.Store{api.dataStore.Backend}
	dimension := dimension.Store{api.dataStore.Backend}
	api.router.HandleFunc("/instances", instance.GetList).Methods("GET")
	api.router.HandleFunc("/instances", api.privateAuth.Check(instance.Add)).Methods("POST")
	api.router.HandleFunc("/instances/{id}", instance.Get).Methods("GET")
	api.router.HandleFunc("/instances/{id}", api.privateAuth.Check(instance.Update)).Methods("PUT")
	api.router.HandleFunc("/instances/{id}/events", api.privateAuth.Check(instance.AddEvent)).Methods("POST")
	api.router.HandleFunc("/instances/{id}/dimensions", api.privateAuth.Check(dimension.GetNodes)).Methods("GET")
	api.router.HandleFunc("/instances/{id}/dimensions/{dimension}/options", dimension.GetUnique).Methods("GET")
	api.router.HandleFunc("/instances/{id}/dimensions/{dimension}/options/{value}", api.privateAuth.Check(instance.AddDimension)).Methods("PUT")
	api.router.HandleFunc("/instances/{id}/dimensions/{dimension}/options/{value}/node_id/{node_id}", api.privateAuth.Check(dimension.AddNodeID)).Methods("PUT")
	api.router.HandleFunc("/instances/{id}/inserted_observations/{inserted_observations}", api.privateAuth.Check(instance.UpdateObservations)).Methods("PUT")
	return &api
}
