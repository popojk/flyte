// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	context "context"

	utils "github.com/flyteorg/flyte/flytestdlib/utils"
	mock "github.com/stretchr/testify/mock"
)

// AutoRefreshCache is an autogenerated mock type for the AutoRefreshCache type
type AutoRefreshCache struct {
	mock.Mock
}

type AutoRefreshCache_Expecter struct {
	mock *mock.Mock
}

func (_m *AutoRefreshCache) EXPECT() *AutoRefreshCache_Expecter {
	return &AutoRefreshCache_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: id
func (_m *AutoRefreshCache) Get(id string) utils.CacheItem {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 utils.CacheItem
	if rf, ok := ret.Get(0).(func(string) utils.CacheItem); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(utils.CacheItem)
		}
	}

	return r0
}

// AutoRefreshCache_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type AutoRefreshCache_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - id string
func (_e *AutoRefreshCache_Expecter) Get(id interface{}) *AutoRefreshCache_Get_Call {
	return &AutoRefreshCache_Get_Call{Call: _e.mock.On("Get", id)}
}

func (_c *AutoRefreshCache_Get_Call) Run(run func(id string)) *AutoRefreshCache_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *AutoRefreshCache_Get_Call) Return(_a0 utils.CacheItem) *AutoRefreshCache_Get_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *AutoRefreshCache_Get_Call) RunAndReturn(run func(string) utils.CacheItem) *AutoRefreshCache_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetOrCreate provides a mock function with given fields: item
func (_m *AutoRefreshCache) GetOrCreate(item utils.CacheItem) (utils.CacheItem, error) {
	ret := _m.Called(item)

	if len(ret) == 0 {
		panic("no return value specified for GetOrCreate")
	}

	var r0 utils.CacheItem
	var r1 error
	if rf, ok := ret.Get(0).(func(utils.CacheItem) (utils.CacheItem, error)); ok {
		return rf(item)
	}
	if rf, ok := ret.Get(0).(func(utils.CacheItem) utils.CacheItem); ok {
		r0 = rf(item)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(utils.CacheItem)
		}
	}

	if rf, ok := ret.Get(1).(func(utils.CacheItem) error); ok {
		r1 = rf(item)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AutoRefreshCache_GetOrCreate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOrCreate'
type AutoRefreshCache_GetOrCreate_Call struct {
	*mock.Call
}

// GetOrCreate is a helper method to define mock.On call
//   - item utils.CacheItem
func (_e *AutoRefreshCache_Expecter) GetOrCreate(item interface{}) *AutoRefreshCache_GetOrCreate_Call {
	return &AutoRefreshCache_GetOrCreate_Call{Call: _e.mock.On("GetOrCreate", item)}
}

func (_c *AutoRefreshCache_GetOrCreate_Call) Run(run func(item utils.CacheItem)) *AutoRefreshCache_GetOrCreate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(utils.CacheItem))
	})
	return _c
}

func (_c *AutoRefreshCache_GetOrCreate_Call) Return(_a0 utils.CacheItem, _a1 error) *AutoRefreshCache_GetOrCreate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AutoRefreshCache_GetOrCreate_Call) RunAndReturn(run func(utils.CacheItem) (utils.CacheItem, error)) *AutoRefreshCache_GetOrCreate_Call {
	_c.Call.Return(run)
	return _c
}

// Start provides a mock function with given fields: ctx
func (_m *AutoRefreshCache) Start(ctx context.Context) {
	_m.Called(ctx)
}

// AutoRefreshCache_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type AutoRefreshCache_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
//   - ctx context.Context
func (_e *AutoRefreshCache_Expecter) Start(ctx interface{}) *AutoRefreshCache_Start_Call {
	return &AutoRefreshCache_Start_Call{Call: _e.mock.On("Start", ctx)}
}

func (_c *AutoRefreshCache_Start_Call) Run(run func(ctx context.Context)) *AutoRefreshCache_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *AutoRefreshCache_Start_Call) Return() *AutoRefreshCache_Start_Call {
	_c.Call.Return()
	return _c
}

func (_c *AutoRefreshCache_Start_Call) RunAndReturn(run func(context.Context)) *AutoRefreshCache_Start_Call {
	_c.Call.Return(run)
	return _c
}

// NewAutoRefreshCache creates a new instance of AutoRefreshCache. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAutoRefreshCache(t interface {
	mock.TestingT
	Cleanup(func())
}) *AutoRefreshCache {
	mock := &AutoRefreshCache{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
