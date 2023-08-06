// Code generated by mockery v2.32.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/ergomake/layerform/internal/data/model"
)

// Backend is an autogenerated mock type for the Backend type
type Backend struct {
	mock.Mock
}

type Backend_Expecter struct {
	mock *mock.Mock
}

func (_m *Backend) EXPECT() *Backend_Expecter {
	return &Backend_Expecter{mock: &_m.Mock}
}

// GetLayer provides a mock function with given fields: ctx, name
func (_m *Backend) GetLayer(ctx context.Context, name string) (*model.Layer, error) {
	ret := _m.Called(ctx, name)

	var r0 *model.Layer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Layer, error)); ok {
		return rf(ctx, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Layer); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Layer)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Backend_GetLayer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLayer'
type Backend_GetLayer_Call struct {
	*mock.Call
}

// GetLayer is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *Backend_Expecter) GetLayer(ctx interface{}, name interface{}) *Backend_GetLayer_Call {
	return &Backend_GetLayer_Call{Call: _e.mock.On("GetLayer", ctx, name)}
}

func (_c *Backend_GetLayer_Call) Run(run func(ctx context.Context, name string)) *Backend_GetLayer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Backend_GetLayer_Call) Return(_a0 *model.Layer, _a1 error) *Backend_GetLayer_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Backend_GetLayer_Call) RunAndReturn(run func(context.Context, string) (*model.Layer, error)) *Backend_GetLayer_Call {
	_c.Call.Return(run)
	return _c
}

// ResolveDependencies provides a mock function with given fields: ctx, layer
func (_m *Backend) ResolveDependencies(ctx context.Context, layer *model.Layer) ([]*model.Layer, error) {
	ret := _m.Called(ctx, layer)

	var r0 []*model.Layer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Layer) ([]*model.Layer, error)); ok {
		return rf(ctx, layer)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.Layer) []*model.Layer); ok {
		r0 = rf(ctx, layer)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Layer)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.Layer) error); ok {
		r1 = rf(ctx, layer)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Backend_ResolveDependencies_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ResolveDependencies'
type Backend_ResolveDependencies_Call struct {
	*mock.Call
}

// ResolveDependencies is a helper method to define mock.On call
//   - ctx context.Context
//   - layer *model.Layer
func (_e *Backend_Expecter) ResolveDependencies(ctx interface{}, layer interface{}) *Backend_ResolveDependencies_Call {
	return &Backend_ResolveDependencies_Call{Call: _e.mock.On("ResolveDependencies", ctx, layer)}
}

func (_c *Backend_ResolveDependencies_Call) Run(run func(ctx context.Context, layer *model.Layer)) *Backend_ResolveDependencies_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*model.Layer))
	})
	return _c
}

func (_c *Backend_ResolveDependencies_Call) Return(_a0 []*model.Layer, _a1 error) *Backend_ResolveDependencies_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Backend_ResolveDependencies_Call) RunAndReturn(run func(context.Context, *model.Layer) ([]*model.Layer, error)) *Backend_ResolveDependencies_Call {
	_c.Call.Return(run)
	return _c
}

// NewBackend creates a new instance of Backend. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBackend(t interface {
	mock.TestingT
	Cleanup(func())
}) *Backend {
	mock := &Backend{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
