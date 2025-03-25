// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	context "context"

	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"
)

// AgentMetadataServiceClient is an autogenerated mock type for the AgentMetadataServiceClient type
type AgentMetadataServiceClient struct {
	mock.Mock
}

type AgentMetadataServiceClient_Expecter struct {
	mock *mock.Mock
}

func (_m *AgentMetadataServiceClient) EXPECT() *AgentMetadataServiceClient_Expecter {
	return &AgentMetadataServiceClient_Expecter{mock: &_m.Mock}
}

// GetAgent provides a mock function with given fields: ctx, in, opts
func (_m *AgentMetadataServiceClient) GetAgent(ctx context.Context, in *admin.GetAgentRequest, opts ...grpc.CallOption) (*admin.GetAgentResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetAgent")
	}

	var r0 *admin.GetAgentResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.GetAgentRequest, ...grpc.CallOption) (*admin.GetAgentResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.GetAgentRequest, ...grpc.CallOption) *admin.GetAgentResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.GetAgentResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.GetAgentRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AgentMetadataServiceClient_GetAgent_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAgent'
type AgentMetadataServiceClient_GetAgent_Call struct {
	*mock.Call
}

// GetAgent is a helper method to define mock.On call
//   - ctx context.Context
//   - in *admin.GetAgentRequest
//   - opts ...grpc.CallOption
func (_e *AgentMetadataServiceClient_Expecter) GetAgent(ctx interface{}, in interface{}, opts ...interface{}) *AgentMetadataServiceClient_GetAgent_Call {
	return &AgentMetadataServiceClient_GetAgent_Call{Call: _e.mock.On("GetAgent",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *AgentMetadataServiceClient_GetAgent_Call) Run(run func(ctx context.Context, in *admin.GetAgentRequest, opts ...grpc.CallOption)) *AgentMetadataServiceClient_GetAgent_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*admin.GetAgentRequest), variadicArgs...)
	})
	return _c
}

func (_c *AgentMetadataServiceClient_GetAgent_Call) Return(_a0 *admin.GetAgentResponse, _a1 error) *AgentMetadataServiceClient_GetAgent_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AgentMetadataServiceClient_GetAgent_Call) RunAndReturn(run func(context.Context, *admin.GetAgentRequest, ...grpc.CallOption) (*admin.GetAgentResponse, error)) *AgentMetadataServiceClient_GetAgent_Call {
	_c.Call.Return(run)
	return _c
}

// ListAgents provides a mock function with given fields: ctx, in, opts
func (_m *AgentMetadataServiceClient) ListAgents(ctx context.Context, in *admin.ListAgentsRequest, opts ...grpc.CallOption) (*admin.ListAgentsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListAgents")
	}

	var r0 *admin.ListAgentsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *admin.ListAgentsRequest, ...grpc.CallOption) (*admin.ListAgentsResponse, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *admin.ListAgentsRequest, ...grpc.CallOption) *admin.ListAgentsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.ListAgentsResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *admin.ListAgentsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AgentMetadataServiceClient_ListAgents_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListAgents'
type AgentMetadataServiceClient_ListAgents_Call struct {
	*mock.Call
}

// ListAgents is a helper method to define mock.On call
//   - ctx context.Context
//   - in *admin.ListAgentsRequest
//   - opts ...grpc.CallOption
func (_e *AgentMetadataServiceClient_Expecter) ListAgents(ctx interface{}, in interface{}, opts ...interface{}) *AgentMetadataServiceClient_ListAgents_Call {
	return &AgentMetadataServiceClient_ListAgents_Call{Call: _e.mock.On("ListAgents",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *AgentMetadataServiceClient_ListAgents_Call) Run(run func(ctx context.Context, in *admin.ListAgentsRequest, opts ...grpc.CallOption)) *AgentMetadataServiceClient_ListAgents_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*admin.ListAgentsRequest), variadicArgs...)
	})
	return _c
}

func (_c *AgentMetadataServiceClient_ListAgents_Call) Return(_a0 *admin.ListAgentsResponse, _a1 error) *AgentMetadataServiceClient_ListAgents_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AgentMetadataServiceClient_ListAgents_Call) RunAndReturn(run func(context.Context, *admin.ListAgentsRequest, ...grpc.CallOption) (*admin.ListAgentsResponse, error)) *AgentMetadataServiceClient_ListAgents_Call {
	_c.Call.Return(run)
	return _c
}

// NewAgentMetadataServiceClient creates a new instance of AgentMetadataServiceClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAgentMetadataServiceClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *AgentMetadataServiceClient {
	mock := &AgentMetadataServiceClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
