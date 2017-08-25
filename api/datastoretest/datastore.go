// Code generated by moq; DO NOT EDIT
// github.com/matryer/moq

package datastoretest

import (
	"github.com/ONSdigital/dp-dataset-api/models"
	"sync"
)

var (
	lockDataStoreMockGetAllDatasets sync.RWMutex
	lockDataStoreMockGetDataset     sync.RWMutex
	lockDataStoreMockGetEdition     sync.RWMutex
	lockDataStoreMockGetEditions    sync.RWMutex
)

// DataStoreMock is a mock implementation of DataStore.
//
//     func TestSomethingThatUsesDataStore(t *testing.T) {
//
//         // make and configure a mocked DataStore
//         mockedDataStore := &DataStoreMock{
//             GetAllDatasetsFunc: func() (*models.DatasetResults, error) {
// 	               panic("TODO: mock out the GetAllDatasets method")
//             },
//             GetDatasetFunc: func(id string) (*models.Dataset, error) {
// 	               panic("TODO: mock out the GetDataset method")
//             },
//             GetEditionFunc: func(datasetID string, editionID string) (*models.Edition, error) {
// 	               panic("TODO: mock out the GetEdition method")
//             },
//             GetEditionsFunc: func(id string) (*models.EditionResults, error) {
// 	               panic("TODO: mock out the GetEditions method")
//             },
//         }
//
//         // TODO: use mockedDataStore in code that requires DataStore
//         //       and then make assertions.
//
//     }
type DataStoreMock struct {
	// GetAllDatasetsFunc mocks the GetAllDatasets method.
	GetAllDatasetsFunc func() (*models.DatasetResults, error)

	// GetDatasetFunc mocks the GetDataset method.
	GetDatasetFunc func(id string) (*models.Dataset, error)

	// GetEditionFunc mocks the GetEdition method.
	GetEditionFunc func(datasetID string, editionID string) (*models.Edition, error)

	// GetEditionsFunc mocks the GetEditions method.
	GetEditionsFunc func(id string) (*models.EditionResults, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetAllDatasets holds details about calls to the GetAllDatasets method.
		GetAllDatasets []struct {
		}
		// GetDataset holds details about calls to the GetDataset method.
		GetDataset []struct {
			// Id is the id argument value.
			Id string
		}
		// GetEdition holds details about calls to the GetEdition method.
		GetEdition []struct {
			// DatasetID is the datasetID argument value.
			DatasetID string
			// EditionID is the editionID argument value.
			EditionID string
		}
		// GetEditions holds details about calls to the GetEditions method.
		GetEditions []struct {
			// Id is the id argument value.
			Id string
		}
	}
}

// GetAllDatasets calls GetAllDatasetsFunc.
func (mock *DataStoreMock) GetAllDatasets() (*models.DatasetResults, error) {
	if mock.GetAllDatasetsFunc == nil {
		panic("moq: DataStoreMock.GetAllDatasetsFunc is nil but DataStore.GetAllDatasets was just called")
	}
	callInfo := struct {
	}{}
	lockDataStoreMockGetAllDatasets.Lock()
	mock.calls.GetAllDatasets = append(mock.calls.GetAllDatasets, callInfo)
	lockDataStoreMockGetAllDatasets.Unlock()
	return mock.GetAllDatasetsFunc()
}

// GetAllDatasetsCalls gets all the calls that were made to GetAllDatasets.
// Check the length with:
//     len(mockedDataStore.GetAllDatasetsCalls())
func (mock *DataStoreMock) GetAllDatasetsCalls() []struct {
} {
	var calls []struct {
	}
	lockDataStoreMockGetAllDatasets.RLock()
	calls = mock.calls.GetAllDatasets
	lockDataStoreMockGetAllDatasets.RUnlock()
	return calls
}

// GetDataset calls GetDatasetFunc.
func (mock *DataStoreMock) GetDataset(id string) (*models.Dataset, error) {
	if mock.GetDatasetFunc == nil {
		panic("moq: DataStoreMock.GetDatasetFunc is nil but DataStore.GetDataset was just called")
	}
	callInfo := struct {
		Id string
	}{
		Id: id,
	}
	lockDataStoreMockGetDataset.Lock()
	mock.calls.GetDataset = append(mock.calls.GetDataset, callInfo)
	lockDataStoreMockGetDataset.Unlock()
	return mock.GetDatasetFunc(id)
}

// GetDatasetCalls gets all the calls that were made to GetDataset.
// Check the length with:
//     len(mockedDataStore.GetDatasetCalls())
func (mock *DataStoreMock) GetDatasetCalls() []struct {
	Id string
} {
	var calls []struct {
		Id string
	}
	lockDataStoreMockGetDataset.RLock()
	calls = mock.calls.GetDataset
	lockDataStoreMockGetDataset.RUnlock()
	return calls
}

// GetEdition calls GetEditionFunc.
func (mock *DataStoreMock) GetEdition(datasetID string, editionID string) (*models.Edition, error) {
	if mock.GetEditionFunc == nil {
		panic("moq: DataStoreMock.GetEditionFunc is nil but DataStore.GetEdition was just called")
	}
	callInfo := struct {
		DatasetID string
		EditionID string
	}{
		DatasetID: datasetID,
		EditionID: editionID,
	}
	lockDataStoreMockGetEdition.Lock()
	mock.calls.GetEdition = append(mock.calls.GetEdition, callInfo)
	lockDataStoreMockGetEdition.Unlock()
	return mock.GetEditionFunc(datasetID, editionID)
}

// GetEditionCalls gets all the calls that were made to GetEdition.
// Check the length with:
//     len(mockedDataStore.GetEditionCalls())
func (mock *DataStoreMock) GetEditionCalls() []struct {
	DatasetID string
	EditionID string
} {
	var calls []struct {
		DatasetID string
		EditionID string
	}
	lockDataStoreMockGetEdition.RLock()
	calls = mock.calls.GetEdition
	lockDataStoreMockGetEdition.RUnlock()
	return calls
}

// GetEditions calls GetEditionsFunc.
func (mock *DataStoreMock) GetEditions(id string) (*models.EditionResults, error) {
	if mock.GetEditionsFunc == nil {
		panic("moq: DataStoreMock.GetEditionsFunc is nil but DataStore.GetEditions was just called")
	}
	callInfo := struct {
		Id string
	}{
		Id: id,
	}
	lockDataStoreMockGetEditions.Lock()
	mock.calls.GetEditions = append(mock.calls.GetEditions, callInfo)
	lockDataStoreMockGetEditions.Unlock()
	return mock.GetEditionsFunc(id)
}

// GetEditionsCalls gets all the calls that were made to GetEditions.
// Check the length with:
//     len(mockedDataStore.GetEditionsCalls())
func (mock *DataStoreMock) GetEditionsCalls() []struct {
	Id string
} {
	var calls []struct {
		Id string
	}
	lockDataStoreMockGetEditions.RLock()
	calls = mock.calls.GetEditions
	lockDataStoreMockGetEditions.RUnlock()
	return calls
}
