package interfaces

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery --name=WorkflowExecutionEventWriter --output=../mocks --case=underscore --with-expecter

type WorkflowExecutionEventWriter interface {
	Run()
	Write(workflowExecutionEvent *admin.WorkflowExecutionEventRequest)
}
