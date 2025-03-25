// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mock "github.com/stretchr/testify/mock"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// ExecutableWorkflowNodeStatus is an autogenerated mock type for the ExecutableWorkflowNodeStatus type
type ExecutableWorkflowNodeStatus struct {
	mock.Mock
}

type ExecutableWorkflowNodeStatus_Expecter struct {
	mock *mock.Mock
}

func (_m *ExecutableWorkflowNodeStatus) EXPECT() *ExecutableWorkflowNodeStatus_Expecter {
	return &ExecutableWorkflowNodeStatus_Expecter{mock: &_m.Mock}
}

// GetExecutionError provides a mock function with given fields:
func (_m *ExecutableWorkflowNodeStatus) GetExecutionError() *core.ExecutionError {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetExecutionError")
	}

	var r0 *core.ExecutionError
	if rf, ok := ret.Get(0).(func() *core.ExecutionError); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.ExecutionError)
		}
	}

	return r0
}

// ExecutableWorkflowNodeStatus_GetExecutionError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetExecutionError'
type ExecutableWorkflowNodeStatus_GetExecutionError_Call struct {
	*mock.Call
}

// GetExecutionError is a helper method to define mock.On call
func (_e *ExecutableWorkflowNodeStatus_Expecter) GetExecutionError() *ExecutableWorkflowNodeStatus_GetExecutionError_Call {
	return &ExecutableWorkflowNodeStatus_GetExecutionError_Call{Call: _e.mock.On("GetExecutionError")}
}

func (_c *ExecutableWorkflowNodeStatus_GetExecutionError_Call) Run(run func()) *ExecutableWorkflowNodeStatus_GetExecutionError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ExecutableWorkflowNodeStatus_GetExecutionError_Call) Return(_a0 *core.ExecutionError) *ExecutableWorkflowNodeStatus_GetExecutionError_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ExecutableWorkflowNodeStatus_GetExecutionError_Call) RunAndReturn(run func() *core.ExecutionError) *ExecutableWorkflowNodeStatus_GetExecutionError_Call {
	_c.Call.Return(run)
	return _c
}

// GetWorkflowNodePhase provides a mock function with given fields:
func (_m *ExecutableWorkflowNodeStatus) GetWorkflowNodePhase() v1alpha1.WorkflowNodePhase {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetWorkflowNodePhase")
	}

	var r0 v1alpha1.WorkflowNodePhase
	if rf, ok := ret.Get(0).(func() v1alpha1.WorkflowNodePhase); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(v1alpha1.WorkflowNodePhase)
	}

	return r0
}

// ExecutableWorkflowNodeStatus_GetWorkflowNodePhase_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetWorkflowNodePhase'
type ExecutableWorkflowNodeStatus_GetWorkflowNodePhase_Call struct {
	*mock.Call
}

// GetWorkflowNodePhase is a helper method to define mock.On call
func (_e *ExecutableWorkflowNodeStatus_Expecter) GetWorkflowNodePhase() *ExecutableWorkflowNodeStatus_GetWorkflowNodePhase_Call {
	return &ExecutableWorkflowNodeStatus_GetWorkflowNodePhase_Call{Call: _e.mock.On("GetWorkflowNodePhase")}
}

func (_c *ExecutableWorkflowNodeStatus_GetWorkflowNodePhase_Call) Run(run func()) *ExecutableWorkflowNodeStatus_GetWorkflowNodePhase_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ExecutableWorkflowNodeStatus_GetWorkflowNodePhase_Call) Return(_a0 v1alpha1.WorkflowNodePhase) *ExecutableWorkflowNodeStatus_GetWorkflowNodePhase_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ExecutableWorkflowNodeStatus_GetWorkflowNodePhase_Call) RunAndReturn(run func() v1alpha1.WorkflowNodePhase) *ExecutableWorkflowNodeStatus_GetWorkflowNodePhase_Call {
	_c.Call.Return(run)
	return _c
}

// NewExecutableWorkflowNodeStatus creates a new instance of ExecutableWorkflowNodeStatus. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewExecutableWorkflowNodeStatus(t interface {
	mock.TestingT
	Cleanup(func())
}) *ExecutableWorkflowNodeStatus {
	mock := &ExecutableWorkflowNodeStatus{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
