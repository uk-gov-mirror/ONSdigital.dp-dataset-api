package api

//go:generate moq -out ../mocks/generated_auth_mocks.go -pkg mocks . AuthHandler
//go:generate moq -out ../mocks/mocks.go -pkg mocks . DownloadsGenerator

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-dataset-api/config"
	"github.com/ONSdigital/dp-dataset-api/dimension"
	"github.com/ONSdigital/dp-dataset-api/instance"
	"github.com/ONSdigital/dp-dataset-api/store"
	"github.com/ONSdigital/dp-dataset-api/url"
	dphandlers "github.com/ONSdigital/dp-net/handlers"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

const (
	downloadServiceToken = "X-Download-Service-Token"

	// audit actions
	addDatasetAction    = "addDataset"
	deleteDatasetAction = "deleteDataset"
	getDatasetsAction   = "getDatasets"
	getDatasetAction    = "getDataset"

	getEditionsAction = "getEditions"
	getEditionAction  = "getEdition"

	getVersionsAction      = "getVersions"
	getVersionAction       = "getVersion"
	updateDatasetAction    = "updateDataset"
	updateVersionAction    = "updateVersion"
	associateVersionAction = "associateVersionAction"
	publishVersionAction   = "publishVersion"
	detachVersionAction    = "detachVersion"

	getDimensionsAction       = "getDimensions"
	getDimensionOptionsAction = "getDimensionOptionsAction"
	getMetadataAction         = "getMetadata"

	hasDownloads = "has_downloads"
)

var (
	trueStringified = strconv.FormatBool(true)

	createPermission = auth.Permissions{Create: true}
	readPermission   = auth.Permissions{Read: true}
	updatePermission = auth.Permissions{Update: true}
	deletePermission = auth.Permissions{Delete: true}
)

//API provides an interface for the routes
type API interface {
	CreateDatasetAPI(string, *mux.Router, store.DataStore) *DatasetAPI
}

// DownloadsGenerator pre generates full file downloads for the specified dataset/edition/version
type DownloadsGenerator interface {
	Generate(ctx context.Context, datasetID, instanceID, edition, version string) error
}

// AuthHandler provides authorisation checks on requests
type AuthHandler interface {
	Require(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}

// DatasetAPI manages importing filters against a dataset
type DatasetAPI struct {
	Router                    *mux.Router
	dataStore                 store.DataStore
	urlBuilder                *url.Builder
	host                      string
	downloadServiceToken      string
	EnablePrePublishView      bool
	downloadGenerator         DownloadsGenerator
	enablePrivateEndpoints    bool
	enableDetachDataset       bool
	enableObservationEndpoint bool
	datasetPermissions        AuthHandler
	permissions               AuthHandler
	instancePublishedChecker  *instance.PublishCheck
	versionPublishedChecker   *PublishCheck
}

// Setup creates a new Dataset API instance and register the API routes based on the application configuration.
func Setup(ctx context.Context, cfg *config.Configuration, router *mux.Router, dataStore store.DataStore, urlBuilder *url.Builder, downloadGenerator DownloadsGenerator, datasetPermissions AuthHandler, permissions AuthHandler) *DatasetAPI {

	api := &DatasetAPI{
		dataStore:                 dataStore,
		host:                      cfg.DatasetAPIURL,
		downloadServiceToken:      cfg.DownloadServiceSecretKey,
		EnablePrePublishView:      cfg.EnablePrivateEndpoints,
		Router:                    router,
		urlBuilder:                urlBuilder,
		downloadGenerator:         downloadGenerator,
		enablePrivateEndpoints:    cfg.EnablePrivateEndpoints,
		enableDetachDataset:       cfg.EnableDetachDataset,
		enableObservationEndpoint: cfg.EnableObservationEndpoint,
		datasetPermissions:        datasetPermissions,
		permissions:               permissions,
		versionPublishedChecker:   nil,
		instancePublishedChecker:  nil,
	}

	if api.enablePrivateEndpoints {
		log.Event(ctx, "enabling private endpoints for dataset api", log.INFO)

		api.versionPublishedChecker = &PublishCheck{
			Datastore: api.dataStore.Backend,
		}

		api.instancePublishedChecker = &instance.PublishCheck{
			Datastore: api.dataStore.Backend,
		}

		instanceAPI := &instance.Store{
			Host:                api.host,
			Storer:              api.dataStore.Backend,
			EnableDetachDataset: api.enableDetachDataset,
		}

		dimensionAPI := &dimension.Store{
			Storer: api.dataStore.Backend,
		}

		api.enablePrivateDatasetEndpoints(ctx)
		api.enablePrivateInstancesEndpoints(instanceAPI)
		api.enablePrivateDimensionsEndpoints(dimensionAPI)
	} else {
		log.Event(ctx, "enabling only public endpoints for dataset api", log.INFO)
		api.enablePublicEndpoints(ctx)
	}
	return api
}

// enablePublicEndpoints register only the public GET endpoints.
func (api *DatasetAPI) enablePublicEndpoints(ctx context.Context) {
	api.get("/datasets", api.getDatasets)
	api.get("/datasets/{dataset_id}", api.getDataset)
	api.get("/datasets/{dataset_id}/editions", api.getEditions)
	api.get("/datasets/{dataset_id}/editions/{edition}", api.getEdition)
	api.get("/datasets/{dataset_id}/editions/{edition}/versions", api.getVersions)
	api.get("/datasets/{dataset_id}/editions/{edition}/versions/{version}", api.getVersion)
	api.get("/datasets/{dataset_id}/editions/{edition}/versions/{version}/metadata", api.getMetadata)
	api.get("/datasets/{dataset_id}/editions/{edition}/versions/{version}/dimensions", api.getDimensions)
	api.get("/datasets/{dataset_id}/editions/{edition}/versions/{version}/dimensions/{dimension}/options", api.getDimensionOptions)

	if api.enableObservationEndpoint {
		log.Event(ctx, "enabling observations endpoint", log.INFO)
		api.get("/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", api.getObservations)
	}
}

// enablePrivateDatasetEndpoints register the datasets endpoints with the appropriate authentication and authorisation
// checks required when running the dataset API in publishing (private) mode.
func (api *DatasetAPI) enablePrivateDatasetEndpoints(ctx context.Context) {
	api.get(
		"/datasets",
		api.isAuthorised(readPermission, api.getDatasets),
	)

	api.get(
		"/datasets/{dataset_id}",
		api.isAuthorisedForDatasets(readPermission,
			api.getDataset),
	)

	api.get(
		"/datasets/{dataset_id}/editions",
		api.isAuthorisedForDatasets(readPermission, api.getEditions),
	)

	api.get(
		"/datasets/{dataset_id}/editions/{edition}",
		api.isAuthorisedForDatasets(readPermission,
			api.getEdition),
	)

	api.get(
		"/datasets/{dataset_id}/editions/{edition}/versions",
		api.isAuthorisedForDatasets(readPermission,
			api.getVersions),
	)

	api.get(
		"/datasets/{dataset_id}/editions/{edition}/versions/{version}",
		api.isAuthorisedForDatasets(readPermission,
			api.getVersion),
	)

	api.get(
		"/datasets/{dataset_id}/editions/{edition}/versions/{version}/metadata",
		api.isAuthorisedForDatasets(readPermission,
			api.getMetadata),
	)

	if api.enableObservationEndpoint {
		log.Event(ctx, "enabling observations endpoint", log.INFO)
		api.get(
			"/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations",
			api.isAuthorisedForDatasets(readPermission,
				api.getObservations),
		)
	}

	api.get(
		"/datasets/{dataset_id}/editions/{edition}/versions/{version}/dimensions",
		api.isAuthorisedForDatasets(readPermission,
			api.getDimensions),
	)

	api.get(
		"/datasets/{dataset_id}/editions/{edition}/versions/{version}/dimensions/{dimension}/options",
		api.isAuthorisedForDatasets(readPermission,
			api.getDimensionOptions),
	)

	api.post(
		"/datasets/{dataset_id}",
		api.isAuthenticated(addDatasetAction,
			api.isAuthorisedForDatasets(createPermission,
				api.addDataset)),
	)

	api.put(
		"/datasets/{dataset_id}",
		api.isAuthenticated(updateDatasetAction,
			api.isAuthorisedForDatasets(updatePermission,
				api.putDataset)),
	)

	api.delete(
		"/datasets/{dataset_id}",
		api.isAuthenticated(deleteDatasetAction,
			api.isAuthorisedForDatasets(deletePermission,
				api.deleteDataset)),
	)

	api.put(
		"/datasets/{dataset_id}/editions/{edition}/versions/{version}",
		api.isAuthenticated(updateVersionAction,
			api.isAuthorisedForDatasets(updatePermission,
				api.isVersionPublished(updateVersionAction,
					api.putVersion))),
	)

	if api.enableDetachDataset {
		api.delete(
			"/datasets/{dataset_id}/editions/{edition}/versions/{version}",
			api.isAuthenticated(detachVersionAction,
				api.isAuthorisedForDatasets(deletePermission,
					api.detachVersion)),
		)
	}
}

// enablePrivateInstancesEndpoints register the instance endpoints with the appropriate authentication and authorisation
// checks required when running the dataset API in publishing (private) mode.
func (api *DatasetAPI) enablePrivateInstancesEndpoints(instanceAPI *instance.Store) {
	api.get(
		"/instances",
		api.isAuthenticated(instance.GetInstancesAction,
			api.isAuthorised(readPermission,
				instanceAPI.GetList)),
	)

	api.post(
		"/instances",
		api.isAuthenticated(instance.AddInstanceAction,
			api.isAuthorised(createPermission,
				instanceAPI.Add)),
	)

	api.get(
		"/instances/{instance_id}",
		api.isAuthenticated(instance.GetInstanceAction,
			api.isAuthorised(readPermission,
				instanceAPI.Get)),
	)

	api.put(
		"/instances/{instance_id}",
		api.isAuthenticated(instance.UpdateInstanceAction,
			api.isAuthorised(updatePermission,
				api.isInstancePublished(instance.UpdateInstanceAction,
					instanceAPI.Update))),
	)

	api.put(
		"/instances/{instance_id}/dimensions/{dimension}",
		api.isAuthenticated(instance.UpdateDimensionAction,
			api.isAuthorised(updatePermission,
				api.isInstancePublished(instance.UpdateDimensionAction,
					instanceAPI.UpdateDimension))),
	)

	api.post(
		"/instances/{instance_id}/events",
		api.isAuthenticated(instance.AddInstanceEventAction,
			api.isAuthorised(createPermission,
				instanceAPI.AddEvent)),
	)

	api.put(
		"/instances/{instance_id}/inserted_observations/{inserted_observations}",
		api.isAuthenticated(instance.UpdateInsertedObservationsAction,
			api.isAuthorised(updatePermission,
				api.isInstancePublished(instance.UpdateInsertedObservationsAction,
					instanceAPI.UpdateObservations))),
	)

	api.put(
		"/instances/{instance_id}/import_tasks",
		api.isAuthenticated(instance.UpdateImportTasksAction,
			api.isAuthorised(updatePermission,
				api.isInstancePublished(instance.UpdateImportTasksAction,
					instanceAPI.UpdateImportTask))),
	)
}

// enablePrivateDatasetEndpoints register the dimenions endpoints with the appropriate authentication and authorisation
// checks required when running the dataset API in publishing (private) mode.
func (api *DatasetAPI) enablePrivateDimensionsEndpoints(dimensionAPI *dimension.Store) {
	api.get(
		"/instances/{instance_id}/dimensions",
		api.isAuthenticated(dimension.GetDimensions,
			api.isAuthorised(readPermission,
				dimensionAPI.GetDimensionsHandler)),
	)

	api.post(
		"/instances/{instance_id}/dimensions",
		api.isAuthenticated(dimension.AddDimensionAction,
			api.isAuthorised(createPermission,
				api.isInstancePublished(dimension.AddDimensionAction,
					dimensionAPI.AddHandler))),
	)

	api.get(
		"/instances/{instance_id}/dimensions/{dimension}/options",
		api.isAuthenticated(dimension.GetUniqueDimensionAndOptionsAction,
			api.isAuthorised(readPermission,
				dimensionAPI.GetUniqueDimensionAndOptionsHandler)),
	)

	api.put(
		"/instances/{instance_id}/dimensions/{dimension}/options/{option}/node_id/{node_id}",
		api.isAuthenticated(dimension.UpdateNodeIDAction,
			api.isAuthorised(updatePermission,
				api.isInstancePublished(dimension.UpdateNodeIDAction,
					dimensionAPI.AddNodeIDHandler))),
	)
}

// isAuthenticated wraps a http handler func in another http handler func that checks the caller is authenticated to
// perform the requested action. action is the audit event name, handler is the http.HandlerFunc to wrap in an
// authentication check. The wrapped handler is only called is the caller is authenticated
func (api *DatasetAPI) isAuthenticated(action string, handler http.HandlerFunc) http.HandlerFunc {
	return dphandlers.CheckIdentity(handler)
}

// isAuthorised wraps a http.HandlerFunc another http.HandlerFunc that checks the caller is authorised to perform the
// requested action. required is the permissions required to perform the action, handler is the http.HandlerFunc to
// apply the check to. The wrapped handler is only called if the caller has the required permissions.
func (api *DatasetAPI) isAuthorised(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc {
	return api.permissions.Require(required, handler)
}

// isAuthorised wraps a http.HandlerFunc another http.HandlerFunc that checks the caller is authorised to perform the
// requested datasets action. This authorisation check is specific to datastes. required is the permissions required to
// perform the action, handler is the http.HandlerFunc to apply the check to. The wrapped handler is only called if the
// caller has the required dataset permissions.
func (api *DatasetAPI) isAuthorisedForDatasets(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc {
	return api.datasetPermissions.Require(required, handler)
}

// isInstancePublished wraps a http.HandlerFunc checking the instance state. The wrapped handler is only invoked if the
// requested instance is not in a published state.
func (api *DatasetAPI) isInstancePublished(action string, handler http.HandlerFunc) http.HandlerFunc {
	return api.instancePublishedChecker.Check(handler, action)
}

// isInstancePublished wraps a http.HandlerFunc checking the version state. The wrapped handler is only invoked if the
// requested version is not in a published state.
func (api *DatasetAPI) isVersionPublished(action string, handler http.HandlerFunc) http.HandlerFunc {
	return api.versionPublishedChecker.Check(handler, action)
}

// get register a GET http.HandlerFunc.
func (api *DatasetAPI) get(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods("GET")
}

// get register a PUT http.HandlerFunc.
func (api *DatasetAPI) put(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods("PUT")
}

// get register a POST http.HandlerFunc.
func (api *DatasetAPI) post(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods("POST")
}

// get register a DELETE http.HandlerFunc.
func (api *DatasetAPI) delete(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods("DELETE")
}

func (api *DatasetAPI) authenticate(r *http.Request, logData log.Data) bool {
	var authorised bool

	if api.EnablePrePublishView {
		var hasCallerIdentity, hasUserIdentity bool

		callerIdentity := dprequest.Caller(r.Context())
		if callerIdentity != "" {
			logData["caller_identity"] = callerIdentity
			hasCallerIdentity = true
		}

		userIdentity := dprequest.User(r.Context())
		if userIdentity != "" {
			logData["user_identity"] = userIdentity
			hasUserIdentity = true
		}

		if hasCallerIdentity || hasUserIdentity {
			authorised = true
		}
		logData["authenticated"] = authorised

		return authorised
	}
	return authorised
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
