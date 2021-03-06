// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package toolcacher

import (
	"context"
	"sync"
)

var (
	lockCacherMockClear sync.RWMutex
	lockCacherMockGet   sync.RWMutex
)

// CacherMock is a mock implementation of Cacher.
//
//     func TestSomethingThatUsesCacher(t *testing.T) {
//
//         // make and configure a mocked Cacher
//         mockedCacher := &CacherMock{
//             ClearFunc: func(ctx context.Context) error {
// 	               panic("mock out the Clear method")
//             },
//             GetFunc: func(ctx context.Context, pkgName string, version string) (string, error) {
// 	               panic("mock out the Get method")
//             },
//         }
//
//         // use mockedCacher in code that requires Cacher
//         // and then make assertions.
//
//     }
type CacherMock struct {
	// ClearFunc mocks the Clear method.
	ClearFunc func(ctx context.Context) error

	// GetFunc mocks the Get method.
	GetFunc func(ctx context.Context, pkgName string, version string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Clear holds details about calls to the Clear method.
		Clear []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// Get holds details about calls to the Get method.
		Get []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// PkgName is the pkgName argument value.
			PkgName string
			// Version is the version argument value.
			Version string
		}
	}
}

// Clear calls ClearFunc.
func (mock *CacherMock) Clear(ctx context.Context) error {
	if mock.ClearFunc == nil {
		panic("CacherMock.ClearFunc: method is nil but Cacher.Clear was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	lockCacherMockClear.Lock()
	mock.calls.Clear = append(mock.calls.Clear, callInfo)
	lockCacherMockClear.Unlock()
	return mock.ClearFunc(ctx)
}

// ClearCalls gets all the calls that were made to Clear.
// Check the length with:
//     len(mockedCacher.ClearCalls())
func (mock *CacherMock) ClearCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	lockCacherMockClear.RLock()
	calls = mock.calls.Clear
	lockCacherMockClear.RUnlock()
	return calls
}

// Get calls GetFunc.
func (mock *CacherMock) Get(ctx context.Context, pkgName string, version string) (string, error) {
	if mock.GetFunc == nil {
		panic("CacherMock.GetFunc: method is nil but Cacher.Get was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		PkgName string
		Version string
	}{
		Ctx:     ctx,
		PkgName: pkgName,
		Version: version,
	}
	lockCacherMockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	lockCacherMockGet.Unlock()
	return mock.GetFunc(ctx, pkgName, version)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//     len(mockedCacher.GetCalls())
func (mock *CacherMock) GetCalls() []struct {
	Ctx     context.Context
	PkgName string
	Version string
} {
	var calls []struct {
		Ctx     context.Context
		PkgName string
		Version string
	}
	lockCacherMockGet.RLock()
	calls = mock.calls.Get
	lockCacherMockGet.RUnlock()
	return calls
}
