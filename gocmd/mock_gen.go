// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package gocmd

import (
	"context"
	"io"
	"sync"
)

var (
	lockCommandMockBuild   sync.RWMutex
	lockCommandMockGet     sync.RWMutex
	lockCommandMockList    sync.RWMutex
	lockCommandMockModTidy sync.RWMutex
)

// CommandMock is a mock implementation of Command.
//
//     func TestSomethingThatUsesCommand(t *testing.T) {
//
//         // make and configure a mocked Command
//         mockedCommand := &CommandMock{
//             BuildFunc: func(ctx context.Context, args ...string) error {
// 	               panic("mock out the Build method")
//             },
//             GetFunc: func(ctx context.Context, args ...string) error {
// 	               panic("mock out the Get method")
//             },
//             ListFunc: func(ctx context.Context, args ...string) (io.Reader, error) {
// 	               panic("mock out the List method")
//             },
//             ModTidyFunc: func(ctx context.Context) error {
// 	               panic("mock out the ModTidy method")
//             },
//         }
//
//         // use mockedCommand in code that requires Command
//         // and then make assertions.
//
//     }
type CommandMock struct {
	// BuildFunc mocks the Build method.
	BuildFunc func(ctx context.Context, args ...string) error

	// GetFunc mocks the Get method.
	GetFunc func(ctx context.Context, args ...string) error

	// ListFunc mocks the List method.
	ListFunc func(ctx context.Context, args ...string) (io.Reader, error)

	// ModTidyFunc mocks the ModTidy method.
	ModTidyFunc func(ctx context.Context) error

	// calls tracks calls to the methods.
	calls struct {
		// Build holds details about calls to the Build method.
		Build []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args []string
		}
		// Get holds details about calls to the Get method.
		Get []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args []string
		}
		// List holds details about calls to the List method.
		List []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Args is the args argument value.
			Args []string
		}
		// ModTidy holds details about calls to the ModTidy method.
		ModTidy []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
}

// Build calls BuildFunc.
func (mock *CommandMock) Build(ctx context.Context, args ...string) error {
	if mock.BuildFunc == nil {
		panic("CommandMock.BuildFunc: method is nil but Command.Build was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args []string
	}{
		Ctx:  ctx,
		Args: args,
	}
	lockCommandMockBuild.Lock()
	mock.calls.Build = append(mock.calls.Build, callInfo)
	lockCommandMockBuild.Unlock()
	return mock.BuildFunc(ctx, args...)
}

// BuildCalls gets all the calls that were made to Build.
// Check the length with:
//     len(mockedCommand.BuildCalls())
func (mock *CommandMock) BuildCalls() []struct {
	Ctx  context.Context
	Args []string
} {
	var calls []struct {
		Ctx  context.Context
		Args []string
	}
	lockCommandMockBuild.RLock()
	calls = mock.calls.Build
	lockCommandMockBuild.RUnlock()
	return calls
}

// Get calls GetFunc.
func (mock *CommandMock) Get(ctx context.Context, args ...string) error {
	if mock.GetFunc == nil {
		panic("CommandMock.GetFunc: method is nil but Command.Get was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args []string
	}{
		Ctx:  ctx,
		Args: args,
	}
	lockCommandMockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	lockCommandMockGet.Unlock()
	return mock.GetFunc(ctx, args...)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//     len(mockedCommand.GetCalls())
func (mock *CommandMock) GetCalls() []struct {
	Ctx  context.Context
	Args []string
} {
	var calls []struct {
		Ctx  context.Context
		Args []string
	}
	lockCommandMockGet.RLock()
	calls = mock.calls.Get
	lockCommandMockGet.RUnlock()
	return calls
}

// List calls ListFunc.
func (mock *CommandMock) List(ctx context.Context, args ...string) (io.Reader, error) {
	if mock.ListFunc == nil {
		panic("CommandMock.ListFunc: method is nil but Command.List was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Args []string
	}{
		Ctx:  ctx,
		Args: args,
	}
	lockCommandMockList.Lock()
	mock.calls.List = append(mock.calls.List, callInfo)
	lockCommandMockList.Unlock()
	return mock.ListFunc(ctx, args...)
}

// ListCalls gets all the calls that were made to List.
// Check the length with:
//     len(mockedCommand.ListCalls())
func (mock *CommandMock) ListCalls() []struct {
	Ctx  context.Context
	Args []string
} {
	var calls []struct {
		Ctx  context.Context
		Args []string
	}
	lockCommandMockList.RLock()
	calls = mock.calls.List
	lockCommandMockList.RUnlock()
	return calls
}

// ModTidy calls ModTidyFunc.
func (mock *CommandMock) ModTidy(ctx context.Context) error {
	if mock.ModTidyFunc == nil {
		panic("CommandMock.ModTidyFunc: method is nil but Command.ModTidy was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	lockCommandMockModTidy.Lock()
	mock.calls.ModTidy = append(mock.calls.ModTidy, callInfo)
	lockCommandMockModTidy.Unlock()
	return mock.ModTidyFunc(ctx)
}

// ModTidyCalls gets all the calls that were made to ModTidy.
// Check the length with:
//     len(mockedCommand.ModTidyCalls())
func (mock *CommandMock) ModTidyCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	lockCommandMockModTidy.RLock()
	calls = mock.calls.ModTidy
	lockCommandMockModTidy.RUnlock()
	return calls
}
