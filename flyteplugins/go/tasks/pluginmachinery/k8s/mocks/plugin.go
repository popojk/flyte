// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	context "context"

	client "sigs.k8s.io/controller-runtime/pkg/client"

	core "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"

	k8s "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"

	mock "github.com/stretchr/testify/mock"
)

// Plugin is an autogenerated mock type for the Plugin type
type Plugin struct {
	mock.Mock
}

type Plugin_Expecter struct {
	mock *mock.Mock
}

func (_m *Plugin) EXPECT() *Plugin_Expecter {
	return &Plugin_Expecter{mock: &_m.Mock}
}

// BuildIdentityResource provides a mock function with given fields: ctx, taskCtx
func (_m *Plugin) BuildIdentityResource(ctx context.Context, taskCtx core.TaskExecutionMetadata) (client.Object, error) {
	ret := _m.Called(ctx, taskCtx)

	if len(ret) == 0 {
		panic("no return value specified for BuildIdentityResource")
	}

	var r0 client.Object
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, core.TaskExecutionMetadata) (client.Object, error)); ok {
		return rf(ctx, taskCtx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, core.TaskExecutionMetadata) client.Object); ok {
		r0 = rf(ctx, taskCtx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Object)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, core.TaskExecutionMetadata) error); ok {
		r1 = rf(ctx, taskCtx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Plugin_BuildIdentityResource_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BuildIdentityResource'
type Plugin_BuildIdentityResource_Call struct {
	*mock.Call
}

// BuildIdentityResource is a helper method to define mock.On call
//   - ctx context.Context
//   - taskCtx core.TaskExecutionMetadata
func (_e *Plugin_Expecter) BuildIdentityResource(ctx interface{}, taskCtx interface{}) *Plugin_BuildIdentityResource_Call {
	return &Plugin_BuildIdentityResource_Call{Call: _e.mock.On("BuildIdentityResource", ctx, taskCtx)}
}

func (_c *Plugin_BuildIdentityResource_Call) Run(run func(ctx context.Context, taskCtx core.TaskExecutionMetadata)) *Plugin_BuildIdentityResource_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(core.TaskExecutionMetadata))
	})
	return _c
}

func (_c *Plugin_BuildIdentityResource_Call) Return(_a0 client.Object, _a1 error) *Plugin_BuildIdentityResource_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Plugin_BuildIdentityResource_Call) RunAndReturn(run func(context.Context, core.TaskExecutionMetadata) (client.Object, error)) *Plugin_BuildIdentityResource_Call {
	_c.Call.Return(run)
	return _c
}

// BuildResource provides a mock function with given fields: ctx, taskCtx
func (_m *Plugin) BuildResource(ctx context.Context, taskCtx core.TaskExecutionContext) (client.Object, error) {
	ret := _m.Called(ctx, taskCtx)

	if len(ret) == 0 {
		panic("no return value specified for BuildResource")
	}

	var r0 client.Object
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, core.TaskExecutionContext) (client.Object, error)); ok {
		return rf(ctx, taskCtx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, core.TaskExecutionContext) client.Object); ok {
		r0 = rf(ctx, taskCtx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.Object)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, core.TaskExecutionContext) error); ok {
		r1 = rf(ctx, taskCtx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Plugin_BuildResource_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BuildResource'
type Plugin_BuildResource_Call struct {
	*mock.Call
}

// BuildResource is a helper method to define mock.On call
//   - ctx context.Context
//   - taskCtx core.TaskExecutionContext
func (_e *Plugin_Expecter) BuildResource(ctx interface{}, taskCtx interface{}) *Plugin_BuildResource_Call {
	return &Plugin_BuildResource_Call{Call: _e.mock.On("BuildResource", ctx, taskCtx)}
}

func (_c *Plugin_BuildResource_Call) Run(run func(ctx context.Context, taskCtx core.TaskExecutionContext)) *Plugin_BuildResource_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(core.TaskExecutionContext))
	})
	return _c
}

func (_c *Plugin_BuildResource_Call) Return(_a0 client.Object, _a1 error) *Plugin_BuildResource_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Plugin_BuildResource_Call) RunAndReturn(run func(context.Context, core.TaskExecutionContext) (client.Object, error)) *Plugin_BuildResource_Call {
	_c.Call.Return(run)
	return _c
}

// GetProperties provides a mock function with given fields:
func (_m *Plugin) GetProperties() k8s.PluginProperties {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetProperties")
	}

	var r0 k8s.PluginProperties
	if rf, ok := ret.Get(0).(func() k8s.PluginProperties); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(k8s.PluginProperties)
	}

	return r0
}

// Plugin_GetProperties_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProperties'
type Plugin_GetProperties_Call struct {
	*mock.Call
}

// GetProperties is a helper method to define mock.On call
func (_e *Plugin_Expecter) GetProperties() *Plugin_GetProperties_Call {
	return &Plugin_GetProperties_Call{Call: _e.mock.On("GetProperties")}
}

func (_c *Plugin_GetProperties_Call) Run(run func()) *Plugin_GetProperties_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Plugin_GetProperties_Call) Return(_a0 k8s.PluginProperties) *Plugin_GetProperties_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Plugin_GetProperties_Call) RunAndReturn(run func() k8s.PluginProperties) *Plugin_GetProperties_Call {
	_c.Call.Return(run)
	return _c
}

// GetTaskPhase provides a mock function with given fields: ctx, pluginContext, resource
func (_m *Plugin) GetTaskPhase(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object) (core.PhaseInfo, error) {
	ret := _m.Called(ctx, pluginContext, resource)

	if len(ret) == 0 {
		panic("no return value specified for GetTaskPhase")
	}

	var r0 core.PhaseInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, k8s.PluginContext, client.Object) (core.PhaseInfo, error)); ok {
		return rf(ctx, pluginContext, resource)
	}
	if rf, ok := ret.Get(0).(func(context.Context, k8s.PluginContext, client.Object) core.PhaseInfo); ok {
		r0 = rf(ctx, pluginContext, resource)
	} else {
		r0 = ret.Get(0).(core.PhaseInfo)
	}

	if rf, ok := ret.Get(1).(func(context.Context, k8s.PluginContext, client.Object) error); ok {
		r1 = rf(ctx, pluginContext, resource)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Plugin_GetTaskPhase_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTaskPhase'
type Plugin_GetTaskPhase_Call struct {
	*mock.Call
}

// GetTaskPhase is a helper method to define mock.On call
//   - ctx context.Context
//   - pluginContext k8s.PluginContext
//   - resource client.Object
func (_e *Plugin_Expecter) GetTaskPhase(ctx interface{}, pluginContext interface{}, resource interface{}) *Plugin_GetTaskPhase_Call {
	return &Plugin_GetTaskPhase_Call{Call: _e.mock.On("GetTaskPhase", ctx, pluginContext, resource)}
}

func (_c *Plugin_GetTaskPhase_Call) Run(run func(ctx context.Context, pluginContext k8s.PluginContext, resource client.Object)) *Plugin_GetTaskPhase_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(k8s.PluginContext), args[2].(client.Object))
	})
	return _c
}

func (_c *Plugin_GetTaskPhase_Call) Return(_a0 core.PhaseInfo, _a1 error) *Plugin_GetTaskPhase_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Plugin_GetTaskPhase_Call) RunAndReturn(run func(context.Context, k8s.PluginContext, client.Object) (core.PhaseInfo, error)) *Plugin_GetTaskPhase_Call {
	_c.Call.Return(run)
	return _c
}

// NewPlugin creates a new instance of Plugin. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPlugin(t interface {
	mock.TestingT
	Cleanup(func())
}) *Plugin {
	mock := &Plugin{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
