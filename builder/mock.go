// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package builder

import (
	"context"
	"sync"
)

var (
	lockBuilderMockBuild sync.RWMutex
)

// BuilderMock is a mock implementation of Builder.
//
//     func TestSomethingThatUsesBuilder(t *testing.T) {
//
//         // make and configure a mocked Builder
//         mockedBuilder := &BuilderMock{
//             BuildFunc: func(ctx context.Context, dir string) error {
// 	               panic("TODO: mock out the Build method")
//             },
//         }
//
//         // TODO: use mockedBuilder in code that requires Builder
//         //       and then make assertions.
//
//     }
type BuilderMock struct {
	// BuildFunc mocks the Build method.
	BuildFunc func(ctx context.Context, dir string) error

	// calls tracks calls to the methods.
	calls struct {
		// Build holds details about calls to the Build method.
		Build []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Dir is the dir argument value.
			Dir string
		}
	}
}

// Build calls BuildFunc.
func (mock *BuilderMock) Build(ctx context.Context, dir string) error {
	if mock.BuildFunc == nil {
		panic("BuilderMock.BuildFunc: method is nil but Builder.Build was just called")
	}
	callInfo := struct {
		Ctx context.Context
		Dir string
	}{
		Ctx: ctx,
		Dir: dir,
	}
	lockBuilderMockBuild.Lock()
	mock.calls.Build = append(mock.calls.Build, callInfo)
	lockBuilderMockBuild.Unlock()
	return mock.BuildFunc(ctx, dir)
}

// BuildCalls gets all the calls that were made to Build.
// Check the length with:
//     len(mockedBuilder.BuildCalls())
func (mock *BuilderMock) BuildCalls() []struct {
	Ctx context.Context
	Dir string
} {
	var calls []struct {
		Ctx context.Context
		Dir string
	}
	lockBuilderMockBuild.RLock()
	calls = mock.calls.Build
	lockBuilderMockBuild.RUnlock()
	return calls
}