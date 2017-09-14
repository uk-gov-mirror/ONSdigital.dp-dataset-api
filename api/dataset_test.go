package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	errs "github.com/ONSdigital/dp-dataset-api/apierrors"
	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/store"
	"github.com/ONSdigital/dp-dataset-api/store/datastoretest"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	host      = "http://localhost:22000"
	secretKey = "coffee"
)

var (
	internalError   = errors.New("internal error")
	badRequestError = errors.New("bad request")
	notFoundError   = errors.New("not found")

	datasetPayload           = `{"contact":{"email":"testing@hotmail.com","name":"John Cox","telephone":"01623 456789"},"description":"census","title":"CensusEthnicity","theme":"population","periodicity":"yearly","state":"completed","next_release":"2016-04-04","publisher":{"name":"The office of national statistics","type":"government department","url":"https://www.ons.gov.uk/"}}`
	editionPayload           = `{"edition":"2017","state":"created"}`
	versionPayload           = `{"instance_id":"a1b2c3","edition":"2017","license":"ONS","release_date":"2017-04-04"}`
	versionAssociatedPayload = `{"instance_id":"a1b2c3","edition":"2017","license":"ONS","release_date":"2017-04-04","state":"associated","collection_id":"12345"}`
	versionPublishedPayload  = `{"instance_id":"a1b2c3","edition":"2017","license":"ONS","release_date":"2017-04-04","state":"published","collection_id":"12345"}`
)

func TestGetDatasetsReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetsFunc: func() (*models.DatasetResults, error) {
				return &models.DatasetResults{}, nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetDatasetsCalls()), ShouldEqual, 1)
	})
}

func TestGetDatasetsReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the api cannot connect to datastore return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetsFunc: func() (*models.DatasetResults, error) {
				return nil, internalError
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetDatasetsCalls()), ShouldEqual, 1)
	})
}

func TestGetDatasetReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("When dataset document has a current sub document return status 200", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(id string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{ID: "123", Current: &models.Dataset{ID: "123"}}, nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
	})

	Convey("When dataset document has only a next sub document and request contains valid internal_token return status 200", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456", nil)
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(id string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{ID: "123", Next: &models.Dataset{ID: "123"}}, nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
	})
}

func TestGetDatasetReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the api cannot connect to datastore return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(id string) (*models.DatasetUpdate, error) {
				return nil, internalError
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
	})

	Convey("When dataset document has only a next sub document return status 404", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(id string) (*models.DatasetUpdate, error) {
				return &models.DatasetUpdate{ID: "123", Next: &models.Dataset{ID: "123"}}, nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
	})

	Convey("When there is no dataset document return status 404", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDatasetFunc: func(id string) (*models.DatasetUpdate, error) {
				return nil, errs.DatasetNotFound
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 1)
	})
}

func TestGetEditionsReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionsFunc: func(id, state string) (*models.EditionResults, error) {
				return &models.EditionResults{}, nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetEditionsCalls()), ShouldEqual, 1)
	})
}

func TestGetEditionsReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the api cannot connect to datastore return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionsFunc: func(id, state string) (*models.EditionResults, error) {
				return nil, internalError
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetEditionsCalls()), ShouldEqual, 1)
	})

	Convey("When no editions exist against an existing dataset return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions", nil)
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionsFunc: func(id, state string) (*models.EditionResults, error) {
				return nil, errs.EditionNotFound
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetEditionsCalls()), ShouldEqual, 1)
	})

	Convey("When no published editions exist against a published dataset return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionsFunc: func(id, state string) (*models.EditionResults, error) {
				return nil, errs.EditionNotFound
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetEditionsCalls()), ShouldEqual, 1)
	})
}

func TestGetEditionReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(id, editionID, state string) (*models.Edition, error) {
				return &models.Edition{}, nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 1)
	})
}

func TestGetEditionReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the api cannot connect to datastore return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(id, editionID, state string) (*models.Edition, error) {
				return nil, internalError
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 1)
	})

	Convey("When edition does not exist for a dataset return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678", nil)
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(id, editionID, state string) (*models.Edition, error) {
				return nil, errs.EditionNotFound
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 1)
	})

	Convey("When edition is not published for a dataset return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(id, editionID, state string) (*models.Edition, error) {
				return nil, errs.EditionNotFound
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 1)
	})
}

func TestGetVersionsReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678/versions", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetVersionsFunc: func(datasetID, editionID, state string) (*models.VersionResults, error) {
				return &models.VersionResults{}, nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetVersionsCalls()), ShouldEqual, 1)
	})
}

func TestGetVersionsReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the api cannot connect to datastore return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678/versions", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetVersionsFunc: func(id, editionID, state string) (*models.VersionResults, error) {
				return nil, internalError
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetVersionsCalls()), ShouldEqual, 1)
	})

	Convey("When version does not exist for an edition of a dataset returns status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678/versions", nil)
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetVersionsFunc: func(datasetID, editionID, state string) (*models.VersionResults, error) {
				return nil, errs.VersionNotFound
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetVersionsCalls()), ShouldEqual, 1)
	})

	Convey("When version is not published against an edition of a dataset return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678/versions", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetVersionsFunc: func(datasetID, editionID, state string) (*models.VersionResults, error) {
				return nil, errs.VersionNotFound
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetVersionsCalls()), ShouldEqual, 1)
	})
}

func TestGetVersionReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678/versions/1", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return &models.Version{}, nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
	})
}

func TestGetVersionReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the api cannot connect to datastore return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678/versions/1", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return nil, internalError
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When version does not exist for an edition of a dataset return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678/versions/1", nil)
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return nil, errs.VersionNotFound
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When version is not published for an edition of a dataset return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123-456/editions/678/versions/1", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetVersionFunc: func(datasetID, editionID, version, state string) (*models.Version, error) {
				return nil, errs.VersionNotFound
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(len(mockedDataStore.GetVersionCalls()), ShouldEqual, 1)
	})
}

func TestPostDatasetsReturnsCreated(t *testing.T) {
	t.Parallel()
	Convey("", t, func() {
		var b string
		b = datasetPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			UpsertDatasetFunc: func(id string, datasetDoc *models.DatasetUpdate) error {
				return nil
			},
		}
		mockedDataStore.UpsertDataset("123", &models.DatasetUpdate{})

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(len(mockedDataStore.UpsertDatasetCalls()), ShouldEqual, 2)
	})
}

func TestPostDatasetReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the request contain malformed json a bad request status is returned", t, func() {
		var b string
		b = "{"
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			UpsertDatasetFunc: func(string, *models.DatasetUpdate) error {
				return badRequestError
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(len(mockedDataStore.UpsertDatasetCalls()), ShouldEqual, 0)
	})

	Convey("When the api cannot connect to datastore return an internal server error", t, func() {
		var b string
		b = datasetPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			UpsertDatasetFunc: func(string, *models.DatasetUpdate) error {
				return internalError
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.UpsertDatasetCalls()), ShouldEqual, 1)
	})

	Convey("When the request does not contain a valid internal token return status unauthorised", t, func() {
		var b string
		b = datasetPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets", bytes.NewBufferString(b))
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			UpsertDatasetFunc: func(string, *models.DatasetUpdate) error {
				return nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
		So(len(mockedDataStore.UpsertDatasetCalls()), ShouldEqual, 0)
	})
}

func TestPostEditionReturnsCreated(t *testing.T) {
	t.Parallel()
	Convey("", t, func() {
		var b string
		b = editionPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(string, string, string) (*models.Edition, error) {
				return nil, errs.EditionNotFound
			},
			UpsertEditionFunc: func(string, *models.Edition) error {
				return nil
			},
		}

		mockedDataStore.GetEdition("123", "2017", "created")

		editionDoc := &models.Edition{
			State: "created",
		}
		mockedDataStore.UpsertEdition("2017", editionDoc)

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpsertEditionCalls()), ShouldEqual, 2)
	})
}

func TestPostEditionReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the request contain malformed json a bad request status is returned", t, func() {
		var b string
		b = "{"
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(string, string, string) (*models.Edition, error) {
				return &models.Edition{State: "created"}, nil
			},
			UpsertEditionFunc: func(string, *models.Edition) error {
				return badRequestError
			},
		}
		mockedDataStore.GetEdition("123", "2017", "created")

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpsertEditionCalls()), ShouldEqual, 0)
	})

	Convey("When the api cannot connect to datastore return an internal server error", t, func() {
		var b string
		b = editionPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(string, string, string) (*models.Edition, error) {
				return nil, internalError
			},
			UpsertEditionFunc: func(string, *models.Edition) error {
				return nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpsertEditionCalls()), ShouldEqual, 0)
	})

	Convey("When the request does not contain a valid internal token return status unauthorised", t, func() {
		var b string
		b = datasetPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017", bytes.NewBufferString(b))
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(string, string, string) (*models.Edition, error) {
				return nil, internalError
			},
			UpsertEditionFunc: func(string, *models.Edition) error {
				return nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpsertEditionCalls()), ShouldEqual, 0)
	})

	Convey("When the edition is already published return status forbidden", t, func() {
		var b string
		b = datasetPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(string, string, string) (*models.Edition, error) {
				return &models.Edition{State: publishedState}, nil
			},
			UpsertEditionFunc: func(string, *models.Edition) error {
				return nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusForbidden)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 1)
		So(len(mockedDataStore.UpsertEditionCalls()), ShouldEqual, 0)
	})
}

func TestPostVersionReturnsCreated(t *testing.T) {
	t.Parallel()
	Convey("When the json body does not contain a state", t, func() {
		var b string
		b = versionPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017/versions", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := setUp("created")

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.GetNextVersionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpsertVersionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpdateDatasetWithAssociationCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpdateEditionCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpsertDatasetCalls()), ShouldEqual, 0)
	})

	Convey("When the json body contains a state of associated", t, func() {
		var b string
		b = versionAssociatedPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017/versions", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := setUp(associatedState)

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.GetNextVersionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpsertVersionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpdateDatasetWithAssociationCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpdateEditionCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 0)
		So(len(mockedDataStore.UpsertDatasetCalls()), ShouldEqual, 0)

	})

	Convey("When the json body contains a state of published", t, func() {
		var b string
		b = versionPublishedPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017/versions", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := setUp(publishedState)

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusCreated)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.GetNextVersionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpsertVersionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpdateEditionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.GetDatasetCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpsertDatasetCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpdateDatasetWithAssociationCalls()), ShouldEqual, 0)
	})
}

func TestPostVersionReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the request contain malformed json a bad request status is returned", t, func() {
		var b string
		b = "{"
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017/versions", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			UpsertVersionFunc: func(string, *models.Version) error {
				return badRequestError
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(len(mockedDataStore.UpsertVersionCalls()), ShouldEqual, 0)
	})

	Convey("When the api cannot connect to datastore return an internal server error", t, func() {
		var b string
		b = versionPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017/versions", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(string, string, string) (*models.Edition, error) {
				return nil, internalError
			},
		}
		mockedDataStore.GetEdition("123", "2017", "created")

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 2)
	})

	Convey("When the api cannot connect to datastore to update version resource return an internal server error", t, func() {
		var b string
		b = versionPayload
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017/versions", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(string, string, string) (*models.Edition, error) {
				return &models.Edition{}, nil
			},
			GetNextVersionFunc: func(string, string) (int, error) {
				return 1, nil
			},
			UpsertVersionFunc: func(string, *models.Version) error {
				return internalError
			},
		}

		mockedDataStore.GetEdition("123", "2017", "created")
		mockedDataStore.GetNextVersion("123", "2017")

		versionDoc := &models.Version{
			State:        "published",
			CollectionID: "12345",
		}
		mockedDataStore.UpsertVersion("1", versionDoc)

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.GetNextVersionCalls()), ShouldEqual, 2)
		So(len(mockedDataStore.UpsertVersionCalls()), ShouldEqual, 2)
	})

	Convey("When the request does not contain a valid internal token return status unauthorised", t, func() {
		var b string
		b = `{"edition":"2017","state":"created","license":"ONS","release_date":"2017-04-04","version":"1"}`
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017/versions", bytes.NewBufferString(b))
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(string, string, string) (*models.Edition, error) {
				return &models.Edition{}, nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusUnauthorized)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 0)
	})

	Convey("When a request missing the collection_id to create version document to be associated with a collection return status bad request", t, func() {
		var b string
		b = `{"edition":"2017","state":"associated","license":"ONS","release_date":"2017-04-04"}`
		r := httptest.NewRequest("POST", "http://localhost:22000/datasets/123/editions/2017/versions", bytes.NewBufferString(b))
		r.Header.Add("internal_token", "coffee")
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetEditionFunc: func(string, string, string) (*models.Edition, error) {
				return &models.Edition{}, nil
			},
		}

		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(len(mockedDataStore.GetEditionCalls()), ShouldEqual, 0)
	})
}

func setUp(state string) *storetest.StorerMock {
	mockedDataStore := &storetest.StorerMock{
		GetDatasetFunc: func(string) (*models.DatasetUpdate, error) {
			return &models.DatasetUpdate{ID: "123", Next: &models.Dataset{}}, nil
		},
		GetEditionFunc: func(string, string, string) (*models.Edition, error) {
			return &models.Edition{}, nil
		},
		GetNextVersionFunc: func(string, string) (int, error) {
			return 1, nil
		},
		UpdateDatasetWithAssociationFunc: func(string, string, *models.Version) error {
			return nil
		},
		UpdateEditionFunc: func(string, string) error {
			return nil
		},
		UpsertDatasetFunc: func(string, *models.DatasetUpdate) error {
			return nil
		},
		UpsertVersionFunc: func(string, *models.Version) error {
			return nil
		},
	}

	mockedDataStore.GetEdition("123", "2017", state)

	mockedDataStore.GetNextVersion("123", "2017")

	versionDoc := &models.Version{
		State: state,
	}
	mockedDataStore.UpsertVersion("1", versionDoc)

	if state == publishedState {
		datasetDoc := &models.DatasetUpdate{
			ID: "123",
			Next: &models.Dataset{
				State: state,
			},
		}
		mockedDataStore.UpdateEdition("123", state)
		mockedDataStore.GetDataset("123")
		mockedDataStore.UpsertDataset("123", datasetDoc)
	}

	if state == associatedState {
		versionDoc := &models.Version{
			State: state,
		}
		mockedDataStore.UpdateDatasetWithAssociation("123", state, versionDoc)
	}

	return mockedDataStore
}

func TestGetDimensionsReturnsOk(t *testing.T) {
	t.Parallel()
	Convey("When the request contain valid ids return dimension information", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123/editions/2017/versions/1/dimensions", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDimensionsFunc: func(datasetID, editionID, versionID string) (*models.DatasetDimensionResults, error) {
				return &models.DatasetDimensionResults{}, nil
			},
		}
		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusOK)
	})
}

func TestGetDimensionsReturnsErrors(t *testing.T) {
	t.Parallel()

	Convey("When the request contain invalid ids return not found error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123/editions/2017/versions/1/dimensions", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDimensionsFunc: func(datasetID, editionID, versionID string) (*models.DatasetDimensionResults, error) {
				return nil, errs.DatasetNotFound
			},
		}
		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusNotFound)
	})
	Convey("When the api cannot connect to datastore to get dimension resource return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/123/editions/2017/versions/1/dimensions", nil)
		w := httptest.NewRecorder()
		mockedDataStore := &storetest.StorerMock{
			GetDimensionsFunc: func(datasetID, editionID, versionID string) (*models.DatasetDimensionResults, error) {
				return nil, internalError
			},
		}
		api := CreateDatasetAPI(host, secretKey, mux.NewRouter(), store.DataStore{Backend: mockedDataStore})
		api.router.ServeHTTP(w, r)
		So(w.Code, ShouldEqual, http.StatusInternalServerError)
	})
}
