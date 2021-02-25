// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-dataset-api/service"
	"sync"
)

var (
	lockCloserMockClose sync.RWMutex
)

// Ensure, that CloserMock does implement service.Closer.
// If this is not the case, regenerate this file with moq.
var _ service.Closer = &CloserMock{}

// CloserMock is a mock implementation of service.Closer.
//
//     func TestSomethingThatUsesCloser(t *testing.T) {
//
//         // make and configure a mocked service.Closer
//         mockedCloser := &CloserMock{
//             CloseFunc: func(ctx context.Context) error {
// 	               panic("mock out the Close method")
//             },
//         }
//
//         // use mockedCloser in code that requires service.Closer
//         // and then make assertions.
//
//     }
type CloserMock struct {
	// CloseFunc mocks the Close method.
	CloseFunc func(ctx context.Context) error

	// calls tracks calls to the methods.
	calls struct {
		// Close holds details about calls to the Close method.
		Close []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
}

// Close calls CloseFunc.
func (mock *CloserMock) Close(ctx context.Context) error {
	if mock.CloseFunc == nil {
		panic("CloserMock.CloseFunc: method is nil but Closer.Close was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	lockCloserMockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	lockCloserMockClose.Unlock()
	return mock.CloseFunc(ctx)
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//     len(mockedCloser.CloseCalls())
func (mock *CloserMock) CloseCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	lockCloserMockClose.RLock()
	calls = mock.calls.Close
	lockCloserMockClose.RUnlock()
	return calls
}
