package delete

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flyte/flytectl/cmd/config"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/clusterresourceattribute"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func deleteClusterResourceAttributeSetup() {
	clusterresourceattribute.DefaultDelConfig = &clusterresourceattribute.AttrDeleteConfig{}
	args = []string{}
}

func TestDeleteClusterResourceAttributes(t *testing.T) {
	t.Run("successful project domain attribute deletion commandline", func(t *testing.T) {
		s := testutils.Setup(t)

		deleteClusterResourceAttributeSetup()
		// Empty attribute file
		clusterresourceattribute.DefaultDelConfig.AttrFile = ""
		// No args implying project domain attribute deletion
		s.DeleterExt.EXPECT().DeleteProjectDomainAttributes(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := deleteClusterResourceAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_CLUSTER_RESOURCE)
	})
	t.Run("failed project domain attribute deletion", func(t *testing.T) {
		s := testutils.Setup(t)

		deleteClusterResourceAttributeSetup()
		// No args implying project domain attribute deletion
		s.DeleterExt.EXPECT().DeleteProjectDomainAttributes(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("failed to delete project domain attributes"))
		err := deleteClusterResourceAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to delete project domain attributes"), err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, admin.MatchableResource_CLUSTER_RESOURCE)
	})
	t.Run("successful project domain attribute deletion file", func(t *testing.T) {
		s := testutils.Setup(t)

		deleteClusterResourceAttributeSetup()
		// Empty attribute file
		clusterresourceattribute.DefaultDelConfig.AttrFile = "testdata/valid_project_domain_cluster_attribute.yaml"
		// No args implying project domain attribute deletion
		s.DeleterExt.EXPECT().DeleteProjectDomainAttributes(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		err := deleteClusterResourceAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, "flytesnacks", "development", admin.MatchableResource_CLUSTER_RESOURCE)
	})
	t.Run("successful workflow attribute deletion", func(t *testing.T) {
		s := testutils.Setup(t)

		deleteClusterResourceAttributeSetup()
		// Empty attribute file
		clusterresourceattribute.DefaultDelConfig.AttrFile = ""
		args := []string{"workflow1"}
		s.DeleterExt.EXPECT().DeleteWorkflowAttributes(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deleteClusterResourceAttributes(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow1",
			admin.MatchableResource_CLUSTER_RESOURCE)
	})
	t.Run("failed workflow attribute deletion", func(t *testing.T) {
		s := testutils.Setup(t)

		deleteClusterResourceAttributeSetup()
		// Empty attribute file
		clusterresourceattribute.DefaultDelConfig.AttrFile = ""
		args := []string{"workflow1"}
		s.DeleterExt.EXPECT().DeleteWorkflowAttributes(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(fmt.Errorf("failed to delete workflow attribute"))
		err := deleteClusterResourceAttributes(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("failed to delete workflow attribute"), err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, config.GetConfig().Project, config.GetConfig().Domain, "workflow1",
			admin.MatchableResource_CLUSTER_RESOURCE)
	})
	t.Run("successful workflow attribute deletion file", func(t *testing.T) {
		s := testutils.Setup(t)

		deleteClusterResourceAttributeSetup()
		// Empty attribute file
		clusterresourceattribute.DefaultDelConfig.AttrFile = "testdata/valid_workflow_cluster_attribute.yaml"
		// No args implying project domain attribute deletion
		s.DeleterExt.EXPECT().DeleteWorkflowAttributes(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deleteClusterResourceAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.Nil(t, err)
		s.DeleterExt.AssertCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, "flytesnacks", "development", "core.control_flow.merge_sort.merge_sort",
			admin.MatchableResource_CLUSTER_RESOURCE)
	})
	t.Run("workflow attribute deletion non existent file", func(t *testing.T) {
		s := testutils.Setup(t)

		deleteClusterResourceAttributeSetup()
		// Empty attribute file
		clusterresourceattribute.DefaultDelConfig.AttrFile = testDataNonExistentFile
		// No args implying project domain attribute deletion
		s.DeleterExt.EXPECT().DeleteWorkflowAttributes(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil)
		err := deleteClusterResourceAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		s.DeleterExt.AssertNotCalled(t, "DeleteWorkflowAttributes",
			s.Ctx, "flytesnacks", "development", "core.control_flow.merge_sort.merge_sort",
			admin.MatchableResource_CLUSTER_RESOURCE)
	})
	t.Run("attribute deletion invalid file", func(t *testing.T) {
		s := testutils.Setup(t)

		deleteClusterResourceAttributeSetup()
		// Empty attribute file
		clusterresourceattribute.DefaultDelConfig.AttrFile = testDataInvalidAttrFile
		// No args implying project domain attribute deletion
		err := deleteClusterResourceAttributes(s.Ctx, []string{}, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t,
			"error unmarshaling JSON: while decoding JSON: json: unknown field \"InvalidDomain\"",
			err.Error())
		s.DeleterExt.AssertNotCalled(t, "DeleteProjectDomainAttributes",
			s.Ctx, "flytesnacks", "development", admin.MatchableResource_CLUSTER_RESOURCE)
	})
}
