package instance_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-dataset-api/api"
	errs "github.com/ONSdigital/dp-dataset-api/apierrors"
	"github.com/ONSdigital/dp-dataset-api/config"
	"github.com/ONSdigital/dp-dataset-api/instance"
	"github.com/ONSdigital/dp-dataset-api/mocks"
	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/store"
	storetest "github.com/ONSdigital/dp-dataset-api/store/datastoretest"
	"github.com/ONSdigital/dp-dataset-api/url"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const host = "http://localhost:8080"

var (
	mu          sync.Mutex
	testContext = context.Background()
)

func createRequestWithToken(method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, body)
	ctx := r.Context()
	ctx = dprequest.SetCaller(ctx, "someone@ons.gov.uk")
	r = r.WithContext(ctx)
	return r, err
}

func initAPIWithMockedStore(mockedStore *storetest.StorerMock) *instance.Store {
	api := &instance.Store{
		Storer: mockedStore,
	}
	return api
}

func Test_GetInstancesReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("Given a GET request to retrieve a list of instance resources is made", t, func() {
		Convey("Then return status ok (200)", func() {
			r := httptest.NewRequest("GET", "http://foo/instances", nil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				GetInstancesFunc: func(context.Context, []string, []string, int, int) ([]*models.Instance, int, error) {
					return []*models.Instance{}, 0, nil
				},
			}

			api := initAPIWithMockedStore(mockedDataStore)
			list, totalCount, err := api.GetList(w, r, 20, 0)

			So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
			So(totalCount, ShouldEqual, 0)
			So(list, ShouldResemble, []*models.Instance{})
			So(err, ShouldEqual, nil)
		})

		Convey("When the request includes a filter by state of 'completed' this is delegated to the database function", func() {
			r := httptest.NewRequest("GET", "http://foo/instances?state=completed", nil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				GetInstancesFunc: func(testContext context.Context, state []string, dataset []string, offset, limit int) ([]*models.Instance, int, error) {
					return []*models.Instance{{InstanceID: "test"}}, 1, nil
				},
			}

			api := initAPIWithMockedStore(mockedDataStore)
			list, totalCount, err := api.GetList(w, r, 20, 0)

			So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
			So(mockedDataStore.GetInstancesCalls()[0].States, ShouldResemble, []string{"completed"})
			So(totalCount, ShouldEqual, 1)
			So(list, ShouldResemble, []*models.Instance{{InstanceID: "test"}})
			So(err, ShouldEqual, nil)
		})

		Convey("When the request includes a filter by dataset of 'test' this is delegated to the database function", func() {
			r := httptest.NewRequest("GET", "http://foo/instances?dataset=test", nil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				GetInstancesFunc: func(testContext context.Context, state []string, dataset []string, offset, limit int) ([]*models.Instance, int, error) {
					return []*models.Instance{}, 0, nil
				},
			}

			api := initAPIWithMockedStore(mockedDataStore)
			api.GetList(w, r, 20, 0)

			So(mockedDataStore.GetInstancesCalls()[0].Datasets, ShouldResemble, []string{"test"})
			So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
		})

		Convey("When the request includes a filter by state of multiple values 'completed,edition-confirmed' these are all delegated to the database function", func() {
			r := httptest.NewRequest("GET", "http://foo/instances?state=completed,edition-confirmed", nil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				GetInstancesFunc: func(testContext context.Context, state []string, dataset []string, offset, limit int) ([]*models.Instance, int, error) {
					return []*models.Instance{}, 0, nil
				},
			}

			api := initAPIWithMockedStore(mockedDataStore)
			api.GetList(w, r, 20, 0)

			So(mockedDataStore.GetInstancesCalls()[0].States, ShouldResemble, []string{"completed", "edition-confirmed"})
			So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
		})

		Convey("When the request includes a filter by state of 'completed' and dataset 'test'", func() {
			r := httptest.NewRequest("GET", "http://foo/instances?state=completed&dataset=test", nil)
			w := httptest.NewRecorder()

			mockedDataStore := &storetest.StorerMock{
				GetInstancesFunc: func(testContext context.Context, state []string, dataset []string, offset, limit int) ([]*models.Instance, int, error) {
					return []*models.Instance{}, 0, nil
				},
			}

			api := initAPIWithMockedStore(mockedDataStore)
			api.GetList(w, r, 20, 0)

			So(mockedDataStore.GetInstancesCalls()[0].States, ShouldResemble, []string{"completed"})
			So(mockedDataStore.GetInstancesCalls()[0].Datasets, ShouldResemble, []string{"test"})
			So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
		})
	})
}

func Test_GetInstancesReturnsError(t *testing.T) {
	t.Parallel()
	Convey("Given a GET request to retrieve a list of instance resources is made", t, func() {
		Convey("When the service is unable to connect to the datastore", func() {
			Convey("Then return status internal server error (500)", func() {
				r := httptest.NewRequest("GET", "http://localhost:21800/instances", nil)
				w := httptest.NewRecorder()

				mockedDataStore := &storetest.StorerMock{
					GetInstancesFunc: func(testContext context.Context, state []string, dataset []string, offset, limit int) ([]*models.Instance, int, error) {
						return nil, 0, errs.ErrInternalServer
					},
				}

				api := initAPIWithMockedStore(mockedDataStore)
				api.GetList(w, r, 20, 0)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
				So(len(mockedDataStore.GetInstancesCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the request contains an invalid state to filter on", func() {
			Convey("Then return status bad request (400)", func() {
				r := httptest.NewRequest("GET", "http://foo/instances?state=foo", nil)
				w := httptest.NewRecorder()

				api := initAPIWithMockedStore(&storetest.StorerMock{})
				api.GetList(w, r, 20, 0)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, "bad request - invalid filter state values: [foo]")
			})
		})
	})
}

func Test_GetInstanceReturnsOK(t *testing.T) {
	t.Parallel()
	Convey("Given a GET request to retrieve an instance resource is made", t, func() {
		Convey("When the request is authorised", func() {
			Convey("Then return status ok (200)", func() {
				r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123", nil)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockedDataStore := &storetest.StorerMock{
					GetInstanceFunc: func(ID string) (*models.Instance, error) {
						return &models.Instance{State: models.CreatedState}, nil
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()

				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusOK)
				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)
				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
			})
		})
	})
}

func Test_GetInstanceReturnsError(t *testing.T) {
	t.Parallel()
	Convey("Given a GET request to retrieve an instance resource is made", t, func() {
		Convey("When the service is unable to connect to the datastore", func() {
			Convey("Then return status internal server error (500)", func() {
				r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123", nil)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockedDataStore := &storetest.StorerMock{
					GetInstanceFunc: func(ID string) (*models.Instance, error) {
						return nil, errs.ErrInternalServer
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()

				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)
				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the current instance state is invalid", func() {
			Convey("Then return status internal server error (500)", func() {
				r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123", nil)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockedDataStore := &storetest.StorerMock{
					GetInstanceFunc: func(ID string) (*models.Instance, error) {
						return &models.Instance{State: "gobbledygook"}, nil
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()
				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)
				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the instance resource does not exist", func() {
			Convey("Then return status not found (404)", func() {
				r, err := createRequestWithToken("GET", "http://localhost:21800/instances/123", nil)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockedDataStore := &storetest.StorerMock{
					GetInstanceFunc: func(ID string) (*models.Instance, error) {
						return nil, errs.ErrInstanceNotFound
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()

				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusNotFound)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInstanceNotFound.Error())
				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)
				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
			})
		})
	})
}

type expectedPostInstanceAuditObject struct {
	Action      string
	ContainsKey string
	Result      string
}

func Test_AddInstanceReturnsCreated(t *testing.T) {
	t.Parallel()
	Convey("Given a POST request to create an instance resource", t, func() {
		Convey("When the request is authorised", func() {
			Convey("Then return status created (201)", func() {
				body := strings.NewReader(`{"links": { "job": { "id":"123-456", "href":"http://localhost:2200/jobs/123-456" } } }`)
				r, err := createRequestWithToken("POST", "http://localhost:21800/instances", body)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockedDataStore := &storetest.StorerMock{
					AddInstanceFunc: func(*models.Instance) (*models.Instance, error) {
						return &models.Instance{}, nil
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()
				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusCreated)
				So(len(mockedDataStore.AddInstanceCalls()), ShouldEqual, 1)

				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)
			})
		})
	})
}

func Test_AddInstanceReturnsError(t *testing.T) {
	t.Parallel()
	Convey("Given a POST request to create an instance resources", t, func() {
		Convey("When the service is unable to connect to the datastore", func() {
			Convey("Then return status internal server error (500)", func() {
				body := strings.NewReader(`{"links": {"job": { "id":"123-456", "href":"http://localhost:2200/jobs/123-456" } } }`)
				r, err := createRequestWithToken("POST", "http://localhost:21800/instances", body)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()
				mockedDataStore := &storetest.StorerMock{
					AddInstanceFunc: func(*models.Instance) (*models.Instance, error) {
						return nil, errs.ErrInternalServer
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()
				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)
				So(len(mockedDataStore.AddInstanceCalls()), ShouldEqual, 1)
			})
		})

		Convey("When the request contains invalid json", func() {
			Convey("Then return status bad request (400)", func() {
				body := strings.NewReader(`{`)
				r, err := createRequestWithToken("POST", "http://localhost:21800/instances", body)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockedDataStore := &storetest.StorerMock{
					AddInstanceFunc: func(*models.Instance) (*models.Instance, error) {
						return &models.Instance{}, nil
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()
				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnableToParseJSON.Error())
				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)
				So(len(mockedDataStore.AddInstanceCalls()), ShouldEqual, 0)
			})
		})

		Convey("When the request contains empty json", func() {
			Convey("Then return status bad request (400)", func() {
				body := strings.NewReader(`{}`)
				r, err := createRequestWithToken("POST", "http://localhost:21800/instances", body)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockedDataStore := &storetest.StorerMock{
					AddInstanceFunc: func(*models.Instance) (*models.Instance, error) {
						return &models.Instance{}, nil
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()
				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrMissingJobProperties.Error())
				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)
				So(len(mockedDataStore.AddInstanceCalls()), ShouldEqual, 0)
			})
		})
	})
}

func Test_UpdateInstanceReturnsOk(t *testing.T) {
	t.Parallel()
	Convey("Given a PUT request to update state of an instance resource is made", t, func() {
		Convey("When the requested state change is to 'submitted'", func() {
			Convey("Then return status ok (200)", func() {
				body := strings.NewReader(`{"state":"submitted"}`)
				r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockedDataStore := &storetest.StorerMock{
					GetInstanceFunc: func(id string) (*models.Instance, error) {
						return &models.Instance{
							Links: &models.InstanceLinks{
								Dataset: &models.LinkObject{
									ID:   "234",
									HRef: "example.com/234",
								},
								Self: &models.LinkObject{
									ID:   "123",
									HRef: "example.com/123",
								},
							},
							State: models.CreatedState,
						}, nil
					},
					UpdateInstanceFunc: func(ctx context.Context, id string, i *models.Instance) error {
						return nil
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()
				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusOK)
				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)
				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 3)
				So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
			})
		})

	})
}

func Test_UpdateInstanceReturnsError(t *testing.T) {
	t.Parallel()
	Convey("Given a PUT request to update state of an instance resource is made", t, func() {
		Convey("When the service is unable to connect to the datastore", func() {
			Convey("Then return status internal server error (500)", func() {
				body := strings.NewReader(`{"state":"created"}`)
				r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()

				mockedDataStore := &storetest.StorerMock{
					GetInstanceFunc: func(id string) (*models.Instance, error) {
						return nil, errs.ErrInternalServer
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()
				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())

				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)

				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
			})
		})

		Convey("When the current instance state is invalid", func() {
			Convey("Then return status internal server error (500)", func() {
				r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", strings.NewReader(`{"state":"completed", "edition": "2017"}`))
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()
				mockedDataStore := &storetest.StorerMock{
					GetInstanceFunc: func(id string) (*models.Instance, error) {
						return &models.Instance{State: "gobbledygook"}, nil
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()

				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusInternalServerError)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInternalServer.Error())
				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)
				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
			})
		})

		Convey("When the json body is invalid", func() {
			Convey("Then return status bad request (400)", func() {
				body := strings.NewReader(`{"state":`)
				r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()
				mockedDataStore := &storetest.StorerMock{
					GetInstanceFunc: func(id string) (*models.Instance, error) {
						return &models.Instance{State: "completed"}, nil
					},
				}
				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()

				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrUnableToParseJSON.Error())

				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)

				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
			})
		})

		Convey("When the json body contains fields that are not allowed to be updated", func() {
			Convey("Then return status bad request (400)", func() {
				body := strings.NewReader(`{"links": { "dataset": { "href": "silly-site"}, "version": { "href": "sillier-site"}}}`)
				r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()
				mockedDataStore := &storetest.StorerMock{
					GetInstanceFunc: func(id string) (*models.Instance, error) {
						return &models.Instance{State: "completed"}, nil
					},
					UpdateInstanceFunc: func(ctx context.Context, id string, i *models.Instance) error {
						return nil
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()

				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Body.String(), ShouldContainSubstring, "unable to update instance contains invalid fields: [instance.Links.Dataset instance.Links.Version]")

				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)

				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
			})
		})

		Convey("When the instance does not exist", func() {
			Convey("Then return status not found (404)", func() {
				body := strings.NewReader(`{"edition": "2017"}`)
				r, err := createRequestWithToken("PUT", "http://localhost:21800/instances/123", body)
				So(err, ShouldBeNil)
				w := httptest.NewRecorder()
				mockedDataStore := &storetest.StorerMock{
					GetInstanceFunc: func(id string) (*models.Instance, error) {
						return nil, errs.ErrInstanceNotFound
					},
				}

				datasetPermissions := mocks.NewAuthHandlerMock()
				permissions := mocks.NewAuthHandlerMock()

				datasetAPI := getAPIWithMocks(testContext, mockedDataStore, &mocks.DownloadsGeneratorMock{}, datasetPermissions, permissions)
				datasetAPI.Router.ServeHTTP(w, r)

				So(w.Code, ShouldEqual, http.StatusNotFound)
				So(w.Body.String(), ShouldContainSubstring, errs.ErrInstanceNotFound.Error())

				So(datasetPermissions.Required.Calls, ShouldEqual, 0)
				So(permissions.Required.Calls, ShouldEqual, 1)

				So(len(mockedDataStore.GetInstanceCalls()), ShouldEqual, 1)
				So(len(mockedDataStore.UpdateInstanceCalls()), ShouldEqual, 0)
				So(len(mockedDataStore.AddVersionDetailsToInstanceCalls()), ShouldEqual, 0)
			})
		})
	})
}

var urlBuilder = url.NewBuilder("localhost:20000")

func getAPIWithMocks(ctx context.Context, mockedDataStore store.Storer, mockedGeneratedDownloads api.DownloadsGenerator, datasetPermissions api.AuthHandler, permissions api.AuthHandler) *api.DatasetAPI {
	mu.Lock()
	defer mu.Unlock()
	cfg, err := config.Get()
	So(err, ShouldBeNil)
	cfg.ServiceAuthToken = "dataset"
	cfg.DatasetAPIURL = "http://localhost:22000"
	cfg.EnablePrivateEndpoints = true

	return api.Setup(ctx, cfg, mux.NewRouter(), store.DataStore{Backend: mockedDataStore}, urlBuilder, mockedGeneratedDownloads, datasetPermissions, permissions)
}
