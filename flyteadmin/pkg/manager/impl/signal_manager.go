package impl

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

type signalMetrics struct {
	Scope promutils.Scope
	Set   labeled.Counter
}

type SignalManager struct {
	db      repoInterfaces.Repository
	metrics signalMetrics
}

func getSignalContext(ctx context.Context, identifier *core.SignalIdentifier) context.Context {
	ctx = contextutils.WithProjectDomain(ctx, identifier.GetExecutionId().GetProject(), identifier.GetExecutionId().GetDomain())
	ctx = contextutils.WithWorkflowID(ctx, identifier.GetExecutionId().GetName())
	return contextutils.WithSignalID(ctx, identifier.GetSignalId())
}

func (s *SignalManager) GetOrCreateSignal(ctx context.Context, request *admin.SignalGetOrCreateRequest) (*admin.Signal, error) {
	if err := validation.ValidateSignalGetOrCreateRequest(ctx, request); err != nil {
		logger.Debugf(ctx, "invalid request [%+v]: %v", request, err)
		return nil, err
	}
	ctx = getSignalContext(ctx, request.GetId())

	signalModel, err := transformers.CreateSignalModel(request.GetId(), request.GetType(), nil)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal with id [%+v] and type [+%v] with err: %v", request.GetId(), request.GetType(), err)
		return nil, err
	}

	err = s.db.SignalRepo().GetOrCreate(ctx, &signalModel)
	if err != nil {
		return nil, err
	}

	signal, err := transformers.FromSignalModel(signalModel)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal model [%+v] with err: %v", signalModel, err)
		return nil, err
	}

	return signal, nil
}

func (s *SignalManager) ListSignals(ctx context.Context, request *admin.SignalListRequest) (*admin.SignalList, error) {
	if err := validation.ValidateSignalListRequest(ctx, request); err != nil {
		logger.Debugf(ctx, "ListSignals request [%+v] is invalid: %v", request, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.GetWorkflowExecutionId())

	identifierFilters, err := util.GetWorkflowExecutionIdentifierFilters(ctx, request.GetWorkflowExecutionId(), common.Signal)
	if err != nil {
		return nil, err
	}

	filters, err := util.AddRequestFilters(request.GetFilters(), common.Signal, identifierFilters)
	if err != nil {
		return nil, err
	}

	sortParameter, err := common.NewSortParameter(request.GetSortBy(), models.SignalColumns)
	if err != nil {
		return nil, err
	}

	offset, err := validation.ValidateToken(request.GetToken())
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"invalid pagination token %s for ListSignals", request.GetToken())
	}

	signalModelList, err := s.db.SignalRepo().List(ctx, repoInterfaces.ListResourceInput{
		InlineFilters: filters,
		Offset:        offset,
		Limit:         int(request.GetLimit()),
		SortParameter: sortParameter,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to list signals with request [%+v] with err %v",
			request, err)
		return nil, err
	}

	signalList, err := transformers.FromSignalModels(signalModelList)
	if err != nil {
		logger.Debugf(ctx, "failed to transform signal models for request [%+v] with err: %v", request, err)
		return nil, err
	}
	var token string
	if len(signalList) == int(request.GetLimit()) {
		token = strconv.Itoa(offset + len(signalList))
	}
	return &admin.SignalList{
		Signals: signalList,
		Token:   token,
	}, nil
}

func (s *SignalManager) SetSignal(ctx context.Context, request *admin.SignalSetRequest) (*admin.SignalSetResponse, error) {
	if err := validation.ValidateSignalSetRequest(ctx, s.db, request); err != nil {
		return nil, err
	}
	ctx = getSignalContext(ctx, request.GetId())

	signalModel, err := transformers.CreateSignalModel(request.GetId(), nil, request.GetValue())
	if err != nil {
		logger.Errorf(ctx, "Failed to transform signal with id [%+v] and value [+%v] with err: %v", request.GetId(), request.GetValue(), err)
		return nil, err
	}

	err = s.db.SignalRepo().Update(ctx, signalModel.SignalKey, signalModel.Value)
	if err != nil {
		return nil, err
	}

	s.metrics.Set.Inc(ctx)
	return &admin.SignalSetResponse{}, nil
}

func NewSignalManager(
	db repoInterfaces.Repository,
	scope promutils.Scope) interfaces.SignalInterface {
	metrics := signalMetrics{
		Scope: scope,
		Set:   labeled.NewCounter("num_set", "count of set signals", scope),
	}

	return &SignalManager{
		db:      db,
		metrics: metrics,
	}
}
