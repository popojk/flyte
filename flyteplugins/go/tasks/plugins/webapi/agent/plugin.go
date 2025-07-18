package agent

import (
	"context"
	"encoding/gob"
	"fmt"
	"slices"
	"sync"
	"time"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	flyteIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	pluginErrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/template"
	flyteIO "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const ID = "agent-service"

type AgentService struct {
	mu                 sync.RWMutex
	supportedTaskTypes []string
	CorePlugin         core.Plugin
}

// ContainTaskType check if agent supports this task type.
func (p *AgentService) ContainTaskType(taskType string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return slices.Contains(p.supportedTaskTypes, taskType)
}

// SetSupportedTaskType set supportTaskType in the agent service.
func (p *AgentService) SetSupportedTaskType(taskTypes []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.supportedTaskTypes = taskTypes
}

type Registry map[string]map[int32]*Agent // map[taskTypeName][taskTypeVersion] => Agent

type Plugin struct {
	metricScope promutils.Scope
	cfg         *Config
	cs          *ClientSet
	registry    Registry
	mu          sync.RWMutex
}

type ResourceWrapper struct {
	Phase flyteIdl.TaskExecution_Phase
	// Deprecated: Please Use Phase instead.
	State      admin.State
	Outputs    *flyteIdl.LiteralMap
	Message    string
	LogLinks   []*flyteIdl.TaskLog
	CustomInfo *structpb.Struct
	AgentError *admin.AgentError
}

// IsTerminal is used to avoid making network calls to the agent service if the resource is already in a terminal state.
func (r ResourceWrapper) IsTerminal() bool {
	return r.Phase == flyteIdl.TaskExecution_SUCCEEDED || r.Phase == flyteIdl.TaskExecution_FAILED || r.Phase == flyteIdl.TaskExecution_ABORTED
}

type ResourceMetaWrapper struct {
	OutputPrefix      string
	AgentResourceMeta []byte
	TaskCategory      admin.TaskCategory
}

func (p *Plugin) setRegistry(r Registry) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.registry = r
}

func (p *Plugin) GetConfig() webapi.PluginConfig {
	return GetConfig().WebAPI
}

func (p *Plugin) ResourceRequirements(_ context.Context, _ webapi.TaskExecutionContextReader) (
	namespace core.ResourceNamespace, constraints core.ResourceConstraintsSpec, err error) {

	// Resource requirements are assumed to be the same.
	return "default", p.cfg.ResourceConstraints, nil
}

func (p *Plugin) Create(ctx context.Context, taskCtx webapi.TaskExecutionContextReader) (webapi.ResourceMeta,
	webapi.Resource, error) {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read task template with error: %v", err)
	}
	inputs, err := taskCtx.InputReader().Get(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read inputs with error: %v", err)
	}

	var argTemplate []string
	if taskTemplate.GetContainer() != nil {
		templateParameters := template.Parameters{
			TaskExecMetadata: taskCtx.TaskExecutionMetadata(),
			Inputs:           taskCtx.InputReader(),
			OutputPath:       taskCtx.OutputWriter(),
			Task:             taskCtx.TaskReader(),
		}
		argTemplate = taskTemplate.GetContainer().GetArgs()
		modifiedArgs, err := template.Render(ctx, taskTemplate.GetContainer().GetArgs(), templateParameters)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to render args with error: %v", err)
		}
		taskTemplate.GetContainer().Args = modifiedArgs
		defer func() {
			// Restore unrendered template for subsequent renders.
			taskTemplate.GetContainer().Args = argTemplate
		}()
	}
	outputPrefix := taskCtx.OutputWriter().GetOutputPrefixPath().String()

	taskCategory := admin.TaskCategory{Name: taskTemplate.GetType(), Version: taskTemplate.GetTaskTypeVersion()}
	agent, isSync := p.getFinalAgent(&taskCategory, p.cfg)

	taskExecutionMetadata := buildTaskExecutionMetadata(taskCtx.TaskExecutionMetadata())

	if isSync {
		finalCtx, cancel := getFinalContext(ctx, "ExecuteTaskSync", agent)
		defer cancel()
		client, err := p.getSyncAgentClient(ctx, agent)
		if err != nil {
			return nil, nil, err
		}
		header := &admin.CreateRequestHeader{Template: taskTemplate, OutputPrefix: outputPrefix, TaskExecutionMetadata: &taskExecutionMetadata}
		return p.ExecuteTaskSync(finalCtx, client, header, inputs)
	}

	finalCtx, cancel := getFinalContext(ctx, "CreateTask", agent)
	defer cancel()

	// Use async agent client
	client, err := p.getAsyncAgentClient(ctx, agent)
	if err != nil {
		return nil, nil, err
	}
	request := &admin.CreateTaskRequest{Inputs: inputs, Template: taskTemplate, OutputPrefix: outputPrefix, TaskExecutionMetadata: &taskExecutionMetadata}
	res, err := client.CreateTask(finalCtx, request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create task from agent with %v", err)
	}

	return ResourceMetaWrapper{
		OutputPrefix:      outputPrefix,
		AgentResourceMeta: res.GetResourceMeta(),
		TaskCategory:      taskCategory,
	}, nil, nil
}

func (p *Plugin) ExecuteTaskSync(
	ctx context.Context,
	client service.SyncAgentServiceClient,
	header *admin.CreateRequestHeader,
	inputs *flyteIdl.LiteralMap,
) (webapi.ResourceMeta, webapi.Resource, error) {
	stream, err := client.ExecuteTaskSync(ctx)
	if err != nil {
		logger.Errorf(ctx, "failed to execute task from agent with %v", err)
		return nil, nil, fmt.Errorf("failed to execute task from agent with %v", err)
	}

	headerProto := &admin.ExecuteTaskSyncRequest{
		Part: &admin.ExecuteTaskSyncRequest_Header{
			Header: header,
		},
	}

	err = stream.Send(headerProto)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send headerProto with error: %w", err)
	}
	inputsProto := &admin.ExecuteTaskSyncRequest{
		Part: &admin.ExecuteTaskSyncRequest_Inputs{
			Inputs: inputs,
		},
	}
	err = stream.Send(inputsProto)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to send inputsProto with error: %w", err)
	}

	// Client is done with sending
	if err := stream.CloseSend(); err != nil {
		logger.Errorf(ctx, "failed to close stream with err %s", err.Error())
		return nil, nil, err
	}

	in, err := stream.Recv()
	if err != nil {
		logger.Errorf(ctx, "failed to receive stream from server %s", err.Error())
		return nil, nil, fmt.Errorf("failed to receive stream from server %w", err)
	}
	if in.GetHeader() == nil {
		return nil, nil, fmt.Errorf("expected header in the response, but got none")
	}
	// TODO: Read the streaming output from the agent, and merge it into the final output.
	// For now, Propeller assumes that the output is always in the header.
	resource := in.GetHeader().GetResource()

	return nil, ResourceWrapper{
		Phase:      resource.GetPhase(),
		Outputs:    resource.GetOutputs(),
		Message:    resource.GetMessage(),
		LogLinks:   resource.GetLogLinks(),
		CustomInfo: resource.GetCustomInfo(),
		AgentError: resource.GetAgentError(),
	}, nil
}

func (p *Plugin) Get(ctx context.Context, taskCtx webapi.GetContext) (latest webapi.Resource, err error) {
	metadata := taskCtx.ResourceMeta().(ResourceMetaWrapper)
	agent, _ := p.getFinalAgent(&metadata.TaskCategory, p.cfg)

	client, err := p.getAsyncAgentClient(ctx, agent)
	if err != nil {
		return nil, err
	}
	finalCtx, cancel := getFinalContext(ctx, "GetTask", agent)
	defer cancel()

	request := &admin.GetTaskRequest{
		TaskType:     metadata.TaskCategory.GetName(),
		TaskCategory: &metadata.TaskCategory,
		ResourceMeta: metadata.AgentResourceMeta,
		OutputPrefix: metadata.OutputPrefix,
	}
	res, err := client.GetTask(finalCtx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get task from agent with %v", err)
	}

	return ResourceWrapper{
		Phase:      res.GetResource().GetPhase(),
		State:      res.GetResource().GetState(),
		Outputs:    res.GetResource().GetOutputs(),
		Message:    res.GetResource().GetMessage(),
		LogLinks:   res.GetResource().GetLogLinks(),
		CustomInfo: res.GetResource().GetCustomInfo(),
	}, nil
}

func (p *Plugin) Delete(ctx context.Context, taskCtx webapi.DeleteContext) error {
	if taskCtx.ResourceMeta() == nil {
		return nil
	}
	metadata := taskCtx.ResourceMeta().(ResourceMetaWrapper)
	agent, _ := p.getFinalAgent(&metadata.TaskCategory, p.cfg)

	client, err := p.getAsyncAgentClient(ctx, agent)
	if err != nil {
		return err
	}
	finalCtx, cancel := getFinalContext(ctx, "DeleteTask", agent)
	defer cancel()

	request := &admin.DeleteTaskRequest{
		TaskType:     metadata.TaskCategory.GetName(),
		TaskCategory: &metadata.TaskCategory,
		ResourceMeta: metadata.AgentResourceMeta,
	}
	_, err = client.DeleteTask(finalCtx, request)
	if err != nil {
		return fmt.Errorf("failed to delete task from agent with %v", err)
	}
	return nil
}

func (p *Plugin) Status(ctx context.Context, taskCtx webapi.StatusContext) (phase core.PhaseInfo, err error) {
	resource := taskCtx.Resource().(ResourceWrapper)
	taskInfo := &core.TaskInfo{Logs: resource.LogLinks, CustomInfo: resource.CustomInfo}
	errorCode := pluginErrors.TaskFailedWithError
	if resource.AgentError != nil && resource.AgentError.GetCode() != "" {
		errorCode = resource.AgentError.GetCode()
	}

	switch resource.Phase {
	case flyteIdl.TaskExecution_QUEUED:
		return core.PhaseInfoQueuedWithTaskInfo(time.Now(), core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_WAITING_FOR_RESOURCES:
		return core.PhaseInfoWaitingForResourcesInfo(time.Now(), core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_INITIALIZING:
		return core.PhaseInfoInitializing(time.Now(), core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case flyteIdl.TaskExecution_RUNNING:
		return core.PhaseInfoRunning(core.DefaultPhaseVersion, taskInfo), nil
	case flyteIdl.TaskExecution_SUCCEEDED:
		err = writeOutput(ctx, taskCtx, resource.Outputs)
		if err != nil {
			logger.Errorf(ctx, "failed to write output with err %s", err.Error())
			return core.PhaseInfoUndefined, err
		}
		return core.PhaseInfoSuccess(taskInfo), nil
	case flyteIdl.TaskExecution_ABORTED:
		return core.PhaseInfoFailure(errorCode, "failed to run the job with aborted phase.", taskInfo), nil
	case flyteIdl.TaskExecution_FAILED:
		return core.PhaseInfoFailure(errorCode, fmt.Sprintf("failed to run the job: %s", resource.Message), taskInfo), nil
	case flyteIdl.TaskExecution_RETRYABLE_FAILED:
		return core.PhaseInfoRetryableFailure(errorCode, fmt.Sprintf("failed to run the job: %s", resource.Message), taskInfo), nil
	}
	// The default phase is undefined.
	if resource.Phase != flyteIdl.TaskExecution_UNDEFINED {
		return core.PhaseInfoUndefined, pluginErrors.Errorf(core.SystemErrorCode, "unknown execution phase [%v].", resource.Phase)
	}

	// If the phase is undefined, we will use state to determine the phase.
	switch resource.State {
	case admin.State_PENDING:
		return core.PhaseInfoInitializing(time.Now(), core.DefaultPhaseVersion, resource.Message, taskInfo), nil
	case admin.State_RUNNING:
		return core.PhaseInfoRunning(core.DefaultPhaseVersion, taskInfo), nil
	case admin.State_PERMANENT_FAILURE:
		return core.PhaseInfoFailure(pluginErrors.TaskFailedWithError, "failed to run the job.\n"+resource.Message, taskInfo), nil
	case admin.State_RETRYABLE_FAILURE:
		return core.PhaseInfoRetryableFailure(pluginErrors.TaskFailedWithError, "failed to run the job.\n"+resource.Message, taskInfo), nil
	case admin.State_SUCCEEDED:
		err = writeOutput(ctx, taskCtx, resource.Outputs)
		if err != nil {
			logger.Errorf(ctx, "failed to write output with err %s", err.Error())
			return core.PhaseInfoUndefined, fmt.Errorf("failed to write output with err %s", err.Error())
		}
		return core.PhaseInfoSuccess(taskInfo), nil
	}
	return core.PhaseInfoUndefined, pluginErrors.Errorf(core.SystemErrorCode, "unknown execution state [%v].", resource.State)
}

func (p *Plugin) getSyncAgentClient(ctx context.Context, agent *Deployment) (service.SyncAgentServiceClient, error) {
	client, ok := p.cs.syncAgentClients[agent.Endpoint]
	if !ok {
		conn, err := getGrpcConnection(ctx, agent)
		if err != nil {
			return nil, fmt.Errorf("failed to get grpc connection with error: %v", err)
		}
		client = service.NewSyncAgentServiceClient(conn)
		p.cs.syncAgentClients[agent.Endpoint] = client
	}
	return client, nil
}

func (p *Plugin) getAsyncAgentClient(ctx context.Context, agent *Deployment) (service.AsyncAgentServiceClient, error) {
	client, ok := p.cs.asyncAgentClients[agent.Endpoint]
	if !ok {
		conn, err := getGrpcConnection(ctx, agent)
		if err != nil {
			return nil, fmt.Errorf("failed to get grpc connection with error: %v", err)
		}
		client = service.NewAsyncAgentServiceClient(conn)
		p.cs.asyncAgentClients[agent.Endpoint] = client
	}
	return client, nil
}

func (p *Plugin) watchAgents(ctx context.Context, agentService *AgentService) {
	go wait.Until(func() {
		childCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		clientSet := getAgentClientSets(childCtx)
		agentRegistry := getAgentRegistry(childCtx, clientSet)
		p.setRegistry(agentRegistry)
		agentService.SetSupportedTaskType(maps.Keys(agentRegistry))
	}, p.cfg.PollInterval.Duration, ctx.Done())
}

func (p *Plugin) getFinalAgent(taskCategory *admin.TaskCategory, cfg *Config) (*Deployment, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if agent, exists := p.registry[taskCategory.GetName()][taskCategory.GetVersion()]; exists {
		return agent.AgentDeployment, agent.IsSync
	}
	return &cfg.DefaultAgent, false
}

func writeOutput(ctx context.Context, taskCtx webapi.StatusContext, outputs *flyteIdl.LiteralMap) error {
	taskTemplate, err := taskCtx.TaskReader().Read(ctx)
	if err != nil {
		return err
	}

	if taskTemplate.GetInterface() == nil || taskTemplate.GetInterface().GetOutputs() == nil || taskTemplate.Interface.Outputs.Variables == nil {
		logger.Debugf(ctx, "The task declares no outputs. Skipping writing the outputs.")
		return nil
	}

	var opReader flyteIO.OutputReader
	if outputs != nil {
		logger.Debugf(ctx, "AgentDeployment returned an output.")
		opReader = ioutils.NewInMemoryOutputReader(outputs, nil, nil)
	} else {
		logger.Debugf(ctx, "AgentDeployment didn't return any output, assuming file based outputs.")
		opReader = ioutils.NewRemoteFileOutputReader(ctx, taskCtx.DataStore(), taskCtx.OutputWriter(), 0)
	}
	return taskCtx.OutputWriter().Put(ctx, opReader)
}

func buildTaskExecutionMetadata(taskExecutionMetadata core.TaskExecutionMetadata) admin.TaskExecutionMetadata {
	taskExecutionID := taskExecutionMetadata.GetTaskExecutionID().GetID()

	return admin.TaskExecutionMetadata{
		TaskExecutionId:      &taskExecutionID,
		Namespace:            taskExecutionMetadata.GetNamespace(),
		Labels:               taskExecutionMetadata.GetLabels(),
		Annotations:          taskExecutionMetadata.GetAnnotations(),
		K8SServiceAccount:    taskExecutionMetadata.GetK8sServiceAccount(),
		EnvironmentVariables: taskExecutionMetadata.GetEnvironmentVariables(),
		Identity:             taskExecutionMetadata.GetSecurityContext().RunAs, // nolint:protogetter
	}
}

func newAgentPlugin(agentService *AgentService) webapi.PluginEntry {
	ctx := context.Background()
	cfg := GetConfig()
	clientSet := getAgentClientSets(ctx)
	agentRegistry := getAgentRegistry(ctx, clientSet)
	supportedTaskTypes := maps.Keys(agentRegistry)

	return webapi.PluginEntry{
		ID:                 "agent-service",
		SupportedTaskTypes: supportedTaskTypes,
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			plugin := &Plugin{
				metricScope: iCtx.MetricsScope(),
				cfg:         cfg,
				cs:          clientSet,
				registry:    agentRegistry,
			}
			plugin.watchAgents(ctx, agentService)
			return plugin, nil
		},
	}
}

func RegisterAgentPlugin(agentService *AgentService) {
	gob.Register(ResourceMetaWrapper{})
	gob.Register(ResourceWrapper{})
	pluginmachinery.PluginRegistry().RegisterRemotePlugin(newAgentPlugin(agentService))
}
