package flytek8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	pluginsIOMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	config1 "github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

func dummyTaskExecutionMetadata(resources *v1.ResourceRequirements, extendedResources *core.ExtendedResources, containerImage string, podTemplate *core.K8SPod) pluginsCore.TaskExecutionMetadata {
	taskExecutionMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
	taskExecutionMetadata.On("GetNamespace").Return("test-namespace")
	taskExecutionMetadata.On("GetAnnotations").Return(map[string]string{"annotation-1": "val1"})
	taskExecutionMetadata.On("GetLabels").Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.On("GetOwnerReference").Return(metav1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.On("GetK8sServiceAccount").Return("service-account")
	tID := &pluginsCoreMock.TaskExecutionID{}
	tID.On("GetID").Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.On("GetGeneratedName").Return("some-acceptable-name")
	taskExecutionMetadata.On("GetTaskExecutionID").Return(tID)

	to := &pluginsCoreMock.TaskOverrides{}
	to.On("GetResources").Return(resources)
	to.On("GetExtendedResources").Return(extendedResources)
	to.On("GetContainerImage").Return(containerImage)
	to.On("GetPodTemplate").Return(podTemplate)
	taskExecutionMetadata.On("GetOverrides").Return(to)
	taskExecutionMetadata.On("IsInterruptible").Return(true)
	taskExecutionMetadata.EXPECT().GetPlatformResources().Return(&v1.ResourceRequirements{})
	taskExecutionMetadata.EXPECT().GetEnvironmentVariables().Return(nil)
	taskExecutionMetadata.EXPECT().GetConsoleURL().Return("")
	return taskExecutionMetadata
}

func dummyTaskTemplate() *core.TaskTemplate {
	return &core.TaskTemplate{
		Type: "test",
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Command: []string{"command"},
				Args:    []string{"{{.Input}}"},
			},
		},
	}
}

func dummyInputReader() io.InputReader {
	inputReader := &pluginsIOMock.InputReader{}
	inputReader.EXPECT().GetInputPath().Return(storage.DataReference("test-data-reference"))
	inputReader.EXPECT().GetInputPrefixPath().Return(storage.DataReference("test-data-reference-prefix"))
	inputReader.EXPECT().Get(mock.Anything).Return(&core.LiteralMap{}, nil)
	return inputReader
}

func dummyExecContext(taskTemplate *core.TaskTemplate, r *v1.ResourceRequirements, rm *core.ExtendedResources, containerImage string, podTemplate *core.K8SPod) pluginsCore.TaskExecutionContext {
	ow := &pluginsIOMock.OutputWriter{}
	ow.EXPECT().GetOutputPrefixPath().Return("")
	ow.EXPECT().GetRawOutputPrefix().Return("")
	ow.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	ow.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")

	tCtx := &pluginsCoreMock.TaskExecutionContext{}
	tCtx.EXPECT().TaskExecutionMetadata().Return(dummyTaskExecutionMetadata(r, rm, containerImage, podTemplate))
	tCtx.EXPECT().InputReader().Return(dummyInputReader())
	tCtx.EXPECT().OutputWriter().Return(ow)

	taskReader := &pluginsCoreMock.TaskReader{}
	taskReader.On("Read", mock.Anything).Return(taskTemplate, nil)
	tCtx.EXPECT().TaskReader().Return(taskReader)
	return tCtx
}

func TestPodSetup(t *testing.T) {
	configAccessor := viper.NewAccessor(config1.Options{
		StrictMode:  true,
		SearchPaths: []string{"testdata/config.yaml"},
	})
	err := configAccessor.UpdateConfig(context.TODO())
	assert.NoError(t, err)

	t.Run("ApplyInterruptibleNodeAffinity", TestApplyInterruptibleNodeAffinity)
	t.Run("UpdatePod", updatePod)
	t.Run("ToK8sPodInterruptible", toK8sPodInterruptible)
}

func TestAddRequiredNodeSelectorRequirements(t *testing.T) {
	t.Run("with empty node affinity", func(t *testing.T) {
		affinity := v1.Affinity{}
		nst := v1.NodeSelectorRequirement{
			Key:      "new",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"new"},
		}
		AddRequiredNodeSelectorRequirements(&affinity, nst)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "new",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"new"},
						},
					},
				},
			},
			affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
	})

	t.Run("with existing node affinity", func(t *testing.T) {
		affinity := v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						v1.NodeSelectorTerm{
							MatchExpressions: []v1.NodeSelectorRequirement{
								v1.NodeSelectorRequirement{
									Key:      "required",
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"required"},
								},
							},
						},
					},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
					v1.PreferredSchedulingTerm{
						Weight: 1,
						Preference: v1.NodeSelectorTerm{
							MatchExpressions: []v1.NodeSelectorRequirement{
								v1.NodeSelectorRequirement{
									Key:      "preferred",
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"preferred"},
								},
							},
						},
					},
				},
			},
		}
		nst := v1.NodeSelectorRequirement{
			Key:      "new",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"new"},
		}
		AddRequiredNodeSelectorRequirements(&affinity, nst)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "required",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"required"},
						},
						v1.NodeSelectorRequirement{
							Key:      "new",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"new"},
						},
					},
				},
			},
			affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
		assert.EqualValues(
			t,
			[]v1.PreferredSchedulingTerm{
				v1.PreferredSchedulingTerm{
					Weight: 1,
					Preference: v1.NodeSelectorTerm{
						MatchExpressions: []v1.NodeSelectorRequirement{
							v1.NodeSelectorRequirement{
								Key:      "preferred",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"preferred"},
							},
						},
					},
				},
			},
			affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
		)
	})
}

func TestAddPreferredNodeSelectorRequirements(t *testing.T) {
	t.Run("with empty node affinity", func(t *testing.T) {
		affinity := v1.Affinity{}
		nst := v1.NodeSelectorRequirement{
			Key:      "new",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"new"},
		}
		AddPreferredNodeSelectorRequirements(&affinity, 10, nst)
		assert.EqualValues(
			t,
			[]v1.PreferredSchedulingTerm{
				v1.PreferredSchedulingTerm{
					Weight: 10,
					Preference: v1.NodeSelectorTerm{
						MatchExpressions: []v1.NodeSelectorRequirement{
							v1.NodeSelectorRequirement{
								Key:      "new",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"new"},
							},
						},
					},
				},
			},
			affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
		)
	})

	t.Run("with existing node affinity", func(t *testing.T) {
		affinity := v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						v1.NodeSelectorTerm{
							MatchExpressions: []v1.NodeSelectorRequirement{
								v1.NodeSelectorRequirement{
									Key:      "required",
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"required"},
								},
							},
						},
					},
				},
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
					v1.PreferredSchedulingTerm{
						Weight: 1,
						Preference: v1.NodeSelectorTerm{
							MatchExpressions: []v1.NodeSelectorRequirement{
								v1.NodeSelectorRequirement{
									Key:      "preferred",
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"preferred"},
								},
							},
						},
					},
				},
			},
		}
		nst := v1.NodeSelectorRequirement{
			Key:      "new",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"new"},
		}
		AddPreferredNodeSelectorRequirements(&affinity, 10, nst)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "required",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"required"},
						},
					},
				},
			},
			affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
		assert.EqualValues(
			t,
			[]v1.PreferredSchedulingTerm{
				v1.PreferredSchedulingTerm{
					Weight: 1,
					Preference: v1.NodeSelectorTerm{
						MatchExpressions: []v1.NodeSelectorRequirement{
							v1.NodeSelectorRequirement{
								Key:      "preferred",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"preferred"},
							},
						},
					},
				},
				v1.PreferredSchedulingTerm{
					Weight: 10,
					Preference: v1.NodeSelectorTerm{
						MatchExpressions: []v1.NodeSelectorRequirement{
							v1.NodeSelectorRequirement{
								Key:      "new",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"new"},
							},
						},
					},
				},
			},
			affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
		)
	})
}

func TestApplyInterruptibleNodeAffinity(t *testing.T) {
	t.Run("WithInterruptibleNodeSelectorRequirement", func(t *testing.T) {
		podSpec := v1.PodSpec{}
		ApplyInterruptibleNodeAffinity(true, &podSpec)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "x/interruptible",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
	})

	t.Run("WithNonInterruptibleNodeSelectorRequirement", func(t *testing.T) {
		podSpec := v1.PodSpec{}
		ApplyInterruptibleNodeAffinity(false, &podSpec)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "x/interruptible",
							Operator: v1.NodeSelectorOpDoesNotExist,
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
	})

	t.Run("WithExistingAffinityWithInterruptibleNodeSelectorRequirement", func(t *testing.T) {
		podSpec := v1.PodSpec{
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									v1.NodeSelectorRequirement{
										Key:      "node selector requirement",
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"exists"},
									},
								},
							},
						},
					},
				},
			},
		}
		ApplyInterruptibleNodeAffinity(true, &podSpec)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "node selector requirement",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"exists"},
						},
						v1.NodeSelectorRequirement{
							Key:      "x/interruptible",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
	})
}

func TestApplyExtendedResourcesOverrides(t *testing.T) {
	t4 := &core.ExtendedResources{
		GpuAccelerator: &core.GPUAccelerator{
			Device: "nvidia-tesla-t4",
		},
	}
	partitionedA100 := &core.ExtendedResources{
		GpuAccelerator: &core.GPUAccelerator{
			Device: "nvidia-tesla-a100",
			PartitionSizeValue: &core.GPUAccelerator_PartitionSize{
				PartitionSize: "1g.5gb",
			},
		},
	}
	unpartitionedA100 := &core.ExtendedResources{
		GpuAccelerator: &core.GPUAccelerator{
			Device: "nvidia-tesla-a100",
			PartitionSizeValue: &core.GPUAccelerator_Unpartitioned{
				Unpartitioned: true,
			},
		},
	}

	t.Run("base and overrides are nil", func(t *testing.T) {
		final := applyExtendedResourcesOverrides(nil, nil)
		assert.NotNil(t, final)
	})

	t.Run("base is nil", func(t *testing.T) {
		final := applyExtendedResourcesOverrides(nil, t4)
		assert.EqualValues(
			t,
			t4.GetGpuAccelerator(),
			final.GetGpuAccelerator(),
		)
	})

	t.Run("overrides is nil", func(t *testing.T) {
		final := applyExtendedResourcesOverrides(t4, nil)
		assert.EqualValues(
			t,
			t4.GetGpuAccelerator(),
			final.GetGpuAccelerator(),
		)
	})

	t.Run("merging", func(t *testing.T) {
		final := applyExtendedResourcesOverrides(partitionedA100, unpartitionedA100)
		assert.EqualValues(
			t,
			unpartitionedA100.GetGpuAccelerator(),
			final.GetGpuAccelerator(),
		)
	})
}

func TestApplyGPUNodeSelectors(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		GpuResourceName:           "nvidia.com/gpu",
		GpuDeviceNodeLabel:        "gpu-device",
		GpuPartitionSizeNodeLabel: "gpu-partition-size",
	}))

	basePodSpec := &v1.PodSpec{
		Containers: []v1.Container{
			{
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						"nvidia.com/gpu": resource.MustParse("1"),
					},
				},
			},
		},
	}

	t.Run("without gpu resource", func(t *testing.T) {
		podSpec := &v1.PodSpec{}
		ApplyGPUNodeSelectors(
			podSpec,
			&core.GPUAccelerator{Device: "nvidia-tesla-a100"},
		)
		assert.Nil(t, podSpec.Affinity)
	})

	t.Run("with gpu device spec only", func(t *testing.T) {
		podSpec := basePodSpec.DeepCopy()
		ApplyGPUNodeSelectors(
			podSpec,
			&core.GPUAccelerator{Device: "nvidia-tesla-a100"},
		)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "gpu-device",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-a100"},
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
		assert.EqualValues(
			t,
			[]v1.Toleration{
				{
					Key:      "gpu-device",
					Value:    "nvidia-tesla-a100",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
			},
			podSpec.Tolerations,
		)
	})

	t.Run("with gpu device and partition size spec", func(t *testing.T) {
		podSpec := basePodSpec.DeepCopy()
		ApplyGPUNodeSelectors(
			podSpec,
			&core.GPUAccelerator{
				Device: "nvidia-tesla-a100",
				PartitionSizeValue: &core.GPUAccelerator_PartitionSize{
					PartitionSize: "1g.5gb",
				},
			},
		)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "gpu-device",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-a100"},
						},
						v1.NodeSelectorRequirement{
							Key:      "gpu-partition-size",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"1g.5gb"},
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
		assert.EqualValues(
			t,
			[]v1.Toleration{
				{
					Key:      "gpu-device",
					Value:    "nvidia-tesla-a100",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
				{
					Key:      "gpu-partition-size",
					Value:    "1g.5gb",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
			},
			podSpec.Tolerations,
		)
	})

	t.Run("with unpartitioned gpu device spec", func(t *testing.T) {
		podSpec := basePodSpec.DeepCopy()
		ApplyGPUNodeSelectors(
			podSpec,
			&core.GPUAccelerator{
				Device: "nvidia-tesla-a100",
				PartitionSizeValue: &core.GPUAccelerator_Unpartitioned{
					Unpartitioned: true,
				},
			},
		)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "gpu-device",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-a100"},
						},
						v1.NodeSelectorRequirement{
							Key:      "gpu-partition-size",
							Operator: v1.NodeSelectorOpDoesNotExist,
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
		assert.EqualValues(
			t,
			[]v1.Toleration{
				{
					Key:      "gpu-device",
					Value:    "nvidia-tesla-a100",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
			},
			podSpec.Tolerations,
		)
	})

	t.Run("with unpartitioned gpu device spec with custom node selector and toleration", func(t *testing.T) {
		gpuUnpartitionedNodeSelectorRequirement := v1.NodeSelectorRequirement{
			Key:      "gpu-unpartitioned",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"true"},
		}
		gpuUnpartitionedToleration := v1.Toleration{
			Key:      "gpu-unpartitioned",
			Value:    "true",
			Operator: v1.TolerationOpEqual,
			Effect:   v1.TaintEffectNoSchedule,
		}
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			GpuResourceName:                         "nvidia.com/gpu",
			GpuDeviceNodeLabel:                      "gpu-device",
			GpuPartitionSizeNodeLabel:               "gpu-partition-size",
			GpuUnpartitionedNodeSelectorRequirement: &gpuUnpartitionedNodeSelectorRequirement,
			GpuUnpartitionedToleration:              &gpuUnpartitionedToleration,
		}))

		podSpec := basePodSpec.DeepCopy()
		ApplyGPUNodeSelectors(
			podSpec,
			&core.GPUAccelerator{
				Device: "nvidia-tesla-a100",
				PartitionSizeValue: &core.GPUAccelerator_Unpartitioned{
					Unpartitioned: true,
				},
			},
		)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "gpu-device",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-a100"},
						},
						gpuUnpartitionedNodeSelectorRequirement,
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
		assert.EqualValues(
			t,
			[]v1.Toleration{
				{
					Key:      "gpu-device",
					Value:    "nvidia-tesla-a100",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
				gpuUnpartitionedToleration,
			},
			podSpec.Tolerations,
		)
	})
}

func updatePod(t *testing.T) {
	taskExecutionMetadata := dummyTaskExecutionMetadata(&v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:              resource.MustParse("1024m"),
			v1.ResourceEphemeralStorage: resource.MustParse("100M"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:              resource.MustParse("1024m"),
			v1.ResourceEphemeralStorage: resource.MustParse("100M"),
		},
	}, nil, "", nil)

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Tolerations: []v1.Toleration{
				{
					Key:   "my toleration key",
					Value: "my toleration value",
				},
			},
			NodeSelector: map[string]string{
				"user": "also configured",
			},
		},
	}
	UpdatePod(taskExecutionMetadata, []v1.ResourceRequirements{}, &pod.Spec)
	assert.Equal(t, v1.RestartPolicyNever, pod.Spec.RestartPolicy)
	for _, tol := range pod.Spec.Tolerations {
		if tol.Key == "x/flyte" {
			assert.Equal(t, tol.Value, "interruptible")
			assert.Equal(t, tol.Operator, v1.TolerationOperator("Equal"))
			assert.Equal(t, tol.Effect, v1.TaintEffect("NoSchedule"))
		} else if tol.Key == "my toleration key" {
			assert.Equal(t, tol.Value, "my toleration value")
		} else {
			t.Fatalf("unexpected toleration [%+v]", tol)
		}
	}
	assert.Equal(t, "service-account", pod.Spec.ServiceAccountName)
	assert.Equal(t, "flyte-scheduler", pod.Spec.SchedulerName)
	assert.Len(t, pod.Spec.Tolerations, 2)
	assert.EqualValues(t, map[string]string{
		"x/interruptible": "true",
		"user":            "also configured",
	}, pod.Spec.NodeSelector)
	assert.EqualValues(
		t,
		[]v1.NodeSelectorTerm{
			v1.NodeSelectorTerm{
				MatchExpressions: []v1.NodeSelectorRequirement{
					v1.NodeSelectorRequirement{
						Key:      "x/interruptible",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"true"},
					},
				},
			},
		},
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
	)
}

func TestUpdatePodWithDefaultAffinityAndInterruptibleNodeSelectorRequirement(t *testing.T) {
	taskExecutionMetadata := dummyTaskExecutionMetadata(&v1.ResourceRequirements{}, nil, "", nil)
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		DefaultAffinity: &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						v1.NodeSelectorTerm{
							MatchExpressions: []v1.NodeSelectorRequirement{
								v1.NodeSelectorRequirement{
									Key:      "default node affinity",
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"exists"},
								},
							},
						},
					},
				},
			},
		},
		InterruptibleNodeSelectorRequirement: &v1.NodeSelectorRequirement{
			Key:      "x/interruptible",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"true"},
		},
	}))
	for i := 0; i < 3; i++ {
		podSpec := v1.PodSpec{}
		UpdatePod(taskExecutionMetadata, []v1.ResourceRequirements{}, &podSpec)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "default node affinity",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"exists"},
						},
						v1.NodeSelectorRequirement{
							Key:      "x/interruptible",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
	}
}

func toK8sPodInterruptible(t *testing.T) {
	ctx := context.TODO()

	x := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:              resource.MustParse("1024m"),
			v1.ResourceEphemeralStorage: resource.MustParse("100M"),
			ResourceNvidiaGPU:           resource.MustParse("1"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:              resource.MustParse("1024m"),
			v1.ResourceEphemeralStorage: resource.MustParse("100M"),
		},
	}, nil, "", nil)

	p, _, _, err := ToK8sPodSpec(ctx, x)
	assert.NoError(t, err)
	assert.Len(t, p.Tolerations, 2)
	assert.Equal(t, "x/flyte", p.Tolerations[1].Key)
	assert.Equal(t, "interruptible", p.Tolerations[1].Value)
	assert.Equal(t, 2, len(p.NodeSelector))
	assert.Equal(t, "true", p.NodeSelector["x/interruptible"])
	assert.EqualValues(
		t,
		[]v1.NodeSelectorTerm{
			v1.NodeSelectorTerm{
				MatchExpressions: []v1.NodeSelectorRequirement{
					v1.NodeSelectorRequirement{
						Key:      "x/interruptible",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"true"},
					},
				},
			},
		},
		p.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
	)
}

func TestToK8sPod(t *testing.T) {
	ctx := context.TODO()

	tolGPU := v1.Toleration{
		Key:      "flyte/gpu",
		Value:    "dedicated",
		Operator: v1.TolerationOpEqual,
		Effect:   v1.TaintEffectNoSchedule,
	}

	tolEphemeralStorage := v1.Toleration{
		Key:      "ephemeral-storage",
		Value:    "dedicated",
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectNoSchedule,
	}

	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		ResourceTolerations: map[v1.ResourceName][]v1.Toleration{
			v1.ResourceEphemeralStorage: {tolEphemeralStorage},
			ResourceNvidiaGPU:           {tolGPU},
		},
		DefaultCPURequest:    resource.MustParse("1024m"),
		DefaultMemoryRequest: resource.MustParse("1024Mi"),
	}))

	op := &pluginsIOMock.OutputFilePaths{}
	op.On("GetOutputPrefixPath").Return(storage.DataReference(""))
	op.On("GetRawOutputPrefix").Return(storage.DataReference(""))

	t.Run("WithGPU", func(t *testing.T) {
		x := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:              resource.MustParse("1024m"),
				v1.ResourceEphemeralStorage: resource.MustParse("100M"),
				ResourceNvidiaGPU:           resource.MustParse("1"),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:              resource.MustParse("1024m"),
				v1.ResourceEphemeralStorage: resource.MustParse("100M"),
			},
		}, nil, "", nil)

		p, _, _, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.Equal(t, len(p.Tolerations), 2)
	})

	t.Run("NoGPU", func(t *testing.T) {
		x := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:              resource.MustParse("1024m"),
				v1.ResourceEphemeralStorage: resource.MustParse("100M"),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:              resource.MustParse("1024m"),
				v1.ResourceEphemeralStorage: resource.MustParse("100M"),
			},
		}, nil, "", nil)

		p, _, _, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.Equal(t, len(p.Tolerations), 1)
		assert.Equal(t, "some-acceptable-name", p.Containers[0].Name)
	})

	t.Run("Default toleration, selector, scheduler", func(t *testing.T) {
		x := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:              resource.MustParse("1024m"),
				v1.ResourceEphemeralStorage: resource.MustParse("100M"),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:              resource.MustParse("1024m"),
				v1.ResourceEphemeralStorage: resource.MustParse("100M"),
			},
		}, nil, "", nil)

		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			DefaultNodeSelector: map[string]string{
				"nodeId": "123",
			},
			SchedulerName:        "myScheduler",
			DefaultCPURequest:    resource.MustParse("1024m"),
			DefaultMemoryRequest: resource.MustParse("1024Mi"),
		}))

		p, _, _, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(p.NodeSelector))
		assert.Equal(t, "myScheduler", p.SchedulerName)
		assert.Equal(t, "some-acceptable-name", p.Containers[0].Name)
		assert.Nil(t, p.SecurityContext)
	})

	t.Run("default-pod-sec-ctx", func(t *testing.T) {
		v := int64(1000)
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			DefaultPodSecurityContext: &v1.PodSecurityContext{
				RunAsGroup: &v,
			},
		}))

		x := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{}, nil, "", nil)
		p, _, _, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.NotNil(t, p.SecurityContext)
		assert.Equal(t, *p.SecurityContext.RunAsGroup, v)
	})

	t.Run("enableHostNetwork", func(t *testing.T) {
		enabled := true
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			EnableHostNetworkingPod: &enabled,
		}))
		x := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{}, nil, "", nil)
		p, _, _, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.True(t, p.HostNetwork)
	})

	t.Run("explicitDisableHostNetwork", func(t *testing.T) {
		enabled := false
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			EnableHostNetworkingPod: &enabled,
		}))
		x := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{}, nil, "", nil)
		p, _, _, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.False(t, p.HostNetwork)
	})

	t.Run("skipSettingHostNetwork", func(t *testing.T) {
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{}))
		x := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{}, nil, "", nil)
		p, _, _, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.False(t, p.HostNetwork)
	})

	t.Run("default-pod-dns-config", func(t *testing.T) {
		val1 := "1"
		val2 := "1"
		val3 := "3"
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			DefaultPodDNSConfig: &v1.PodDNSConfig{
				Nameservers: []string{"8.8.8.8", "8.8.4.4"},
				Options: []v1.PodDNSConfigOption{
					{
						Name:  "ndots",
						Value: &val1,
					},
					{
						Name: "single-request-reopen",
					},
					{
						Name:  "timeout",
						Value: &val2,
					},
					{
						Name:  "attempts",
						Value: &val3,
					},
				},
				Searches: []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"},
			},
		}))

		x := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{}, nil, "", nil)
		p, _, _, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.NotNil(t, p.DNSConfig)
		assert.Equal(t, []string{"8.8.8.8", "8.8.4.4"}, p.DNSConfig.Nameservers)
		assert.Equal(t, "ndots", p.DNSConfig.Options[0].Name)
		assert.Equal(t, val1, *p.DNSConfig.Options[0].Value)
		assert.Equal(t, "single-request-reopen", p.DNSConfig.Options[1].Name)
		assert.Equal(t, "timeout", p.DNSConfig.Options[2].Name)
		assert.Equal(t, val2, *p.DNSConfig.Options[2].Value)
		assert.Equal(t, "attempts", p.DNSConfig.Options[3].Name)
		assert.Equal(t, val3, *p.DNSConfig.Options[3].Value)
		assert.Equal(t, []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"}, p.DNSConfig.Searches)
	})

	t.Run("environmentVariables", func(t *testing.T) {
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			DefaultEnvVars: map[string]string{
				"foo": "bar",
			},
		}))
		x := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{}, nil, "", nil)
		p, _, _, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		for _, c := range p.Containers {
			uniqueVariableNames := make(map[string]string)
			for _, envVar := range c.Env {
				if _, ok := uniqueVariableNames[envVar.Name]; ok {
					t.Errorf("duplicate environment variable %s", envVar.Name)
				}
				uniqueVariableNames[envVar.Name] = envVar.Value
			}
		}
	})
}

func TestToK8sPodContainerImage(t *testing.T) {
	t.Run("Override container image", func(t *testing.T) {
		taskContext := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("1024m"),
			}}, nil, "foo:latest", nil)
		p, _, _, err := ToK8sPodSpec(context.TODO(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, "foo:latest", p.Containers[0].Image)
	})
}

func TestPodTemplateOverride(t *testing.T) {
	metadata := &core.K8SObjectMetadata{
		Labels: map[string]string{
			"l": "a",
		},
		Annotations: map[string]string{
			"a": "b",
		},
	}

	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:  "foo",
				Image: "foo:latest",
				Args:  []string{"foo", "bar"},
			},
		},
	}

	podSpecStruct, err := utils.MarshalObjToStruct(podSpec)
	assert.NoError(t, err)

	t.Run("Override pod template", func(t *testing.T) {
		taskContext := dummyExecContext(dummyTaskTemplate(), &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("1024m"),
			}}, nil, "", &core.K8SPod{
			PrimaryContainerName: "foo",
			PodSpec:              podSpecStruct,
			Metadata:             metadata,
		})
		p, m, _, err := ToK8sPodSpec(context.TODO(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, "a", m.Labels["l"])
		assert.Equal(t, "b", m.Annotations["a"])
		assert.Equal(t, "foo:latest", p.Containers[0].Image)
		assert.Equal(t, "foo", p.Containers[0].Name)
	})
}

func TestToK8sPodExtendedResources(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		GpuDeviceNodeLabel:        "gpu-node-label",
		GpuPartitionSizeNodeLabel: "gpu-partition-size",
		GpuResourceName:           ResourceNvidiaGPU,
	}))

	fixtures := []struct {
		name                      string
		resources                 *v1.ResourceRequirements
		extendedResourcesBase     *core.ExtendedResources
		extendedResourcesOverride *core.ExtendedResources
		expectedNsr               []v1.NodeSelectorTerm
		expectedTol               []v1.Toleration
	}{
		{
			"without overrides",
			&v1.ResourceRequirements{
				Limits: v1.ResourceList{
					ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			&core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device: "nvidia-tesla-t4",
				},
			},
			nil,
			[]v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "gpu-node-label",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-t4"},
						},
					},
				},
			},
			[]v1.Toleration{
				{
					Key:      "gpu-node-label",
					Value:    "nvidia-tesla-t4",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
			},
		},
		{
			"with overrides",
			&v1.ResourceRequirements{
				Limits: v1.ResourceList{
					ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			&core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device: "nvidia-tesla-t4",
				},
			},
			&core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device: "nvidia-tesla-a100",
					PartitionSizeValue: &core.GPUAccelerator_PartitionSize{
						PartitionSize: "1g.5gb",
					},
				},
			},
			[]v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "gpu-node-label",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-a100"},
						},
						v1.NodeSelectorRequirement{
							Key:      "gpu-partition-size",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"1g.5gb"},
						},
					},
				},
			},
			[]v1.Toleration{
				{
					Key:      "gpu-node-label",
					Value:    "nvidia-tesla-a100",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
				{
					Key:      "gpu-partition-size",
					Value:    "1g.5gb",
					Operator: v1.TolerationOpEqual,
					Effect:   v1.TaintEffectNoSchedule,
				},
			},
		},
	}

	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			taskTemplate := dummyTaskTemplate()
			taskTemplate.ExtendedResources = f.extendedResourcesBase
			taskContext := dummyExecContext(taskTemplate, f.resources, f.extendedResourcesOverride, "", nil)
			p, _, _, err := ToK8sPodSpec(context.TODO(), taskContext)
			assert.NoError(t, err)

			assert.EqualValues(
				t,
				f.expectedNsr,
				p.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
			)
			assert.EqualValues(
				t,
				f.expectedTol,
				p.Tolerations,
			)
		})
	}
}

func TestDemystifyPending(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		CreateContainerErrorGracePeriod: config1.Duration{
			Duration: time.Minute * 3,
		},
		CreateContainerConfigErrorGracePeriod: config1.Duration{
			Duration: time.Minute * 4,
		},
		ImagePullBackoffGracePeriod: config1.Duration{
			Duration: time.Minute * 3,
		},
		PodPendingTimeout: config1.Duration{
			Duration: 0,
		},
	}))

	t.Run("PodNotScheduled", func(t *testing.T) {
		s := v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodScheduled,
					Status: v1.ConditionFalse,
				},
			},
		}
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseQueued, taskStatus.Phase())
	})

	t.Run("PodUnschedulable", func(t *testing.T) {
		s := v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReasonUnschedulable,
					Status: v1.ConditionFalse,
				},
			},
		}
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseQueued, taskStatus.Phase())
	})

	t.Run("PodNotScheduled", func(t *testing.T) {
		s := v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodScheduled,
					Status: v1.ConditionTrue,
				},
			},
		}
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseQueued, taskStatus.Phase())
	})

	t.Run("PodUnschedulable", func(t *testing.T) {
		s := v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReasonUnschedulable,
					Status: v1.ConditionUnknown,
				},
			},
		}
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseQueued, taskStatus.Phase())
	})

	s := v1.PodStatus{
		Phase: v1.PodPending,
		Conditions: []v1.PodCondition{
			{
				Type:   v1.PodReady,
				Status: v1.ConditionFalse,
			},
			{
				Type:   v1.PodReasonUnschedulable,
				Status: v1.ConditionUnknown,
			},
			{
				Type:   v1.PodScheduled,
				Status: v1.ConditionTrue,
			},
		},
	}

	t.Run("ContainerCreating", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "ContainerCreating",
						Message: "this is not an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseInitializing, taskStatus.Phase())
	})

	t.Run("ErrImagePull", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "ErrImagePull",
						Message: "this is not an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseInitializing, taskStatus.Phase())
	})

	t.Run("PodInitializing", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "PodInitializing",
						Message: "this is not an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseInitializing, taskStatus.Phase())
	})

	t.Run("ImagePullBackOffWithinGracePeriod", func(t *testing.T) {
		s2 := *s.DeepCopy()
		s2.Conditions[0].LastTransitionTime = metav1.Now()
		s2.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "ImagePullBackOff",
						Message: "this is an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s2, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseInitializing, taskStatus.Phase())
	})

	t.Run("ImagePullBackOffOutsideGracePeriod", func(t *testing.T) {
		s2 := *s.DeepCopy()
		s2.Conditions[0].LastTransitionTime.Time = metav1.Now().Add(-config.GetK8sPluginConfig().ImagePullBackoffGracePeriod.Duration)
		s2.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "ImagePullBackOff",
						Message: "this is an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s2, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskStatus.Phase())
		assert.True(t, taskStatus.CleanupOnFailure())
	})

	t.Run("InvalidImageName", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "InvalidImageName",
						Message: "this is an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhasePermanentFailure, taskStatus.Phase())
		assert.True(t, taskStatus.CleanupOnFailure())
	})

	t.Run("RegistryUnavailable", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "RegistryUnavailable",
						Message: "this is an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskStatus.Phase())
		assert.True(t, taskStatus.CleanupOnFailure())
	})

	t.Run("RandomError", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "RandomError",
						Message: "this is an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskStatus.Phase())
		assert.True(t, taskStatus.CleanupOnFailure())
	})

	t.Run("CreateContainerConfigErrorWithinGracePeriod", func(t *testing.T) {
		s2 := *s.DeepCopy()
		s2.Conditions[0].LastTransitionTime = metav1.Now()
		s2.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "CreateContainerConfigError",
						Message: "this is a transient error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s2, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseInitializing, taskStatus.Phase())
	})

	t.Run("CreateContainerConfigErrorOutsideGracePeriod", func(t *testing.T) {
		s2 := *s.DeepCopy()
		s2.Conditions[0].LastTransitionTime.Time = metav1.Now().Add(-config.GetK8sPluginConfig().CreateContainerConfigErrorGracePeriod.Duration)
		s2.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "CreateContainerConfigError",
						Message: "this a permanent error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s2, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhasePermanentFailure, taskStatus.Phase())
		assert.True(t, taskStatus.CleanupOnFailure())
	})

	t.Run("CreateContainerErrorWithinGracePeriod", func(t *testing.T) {
		s2 := *s.DeepCopy()
		s2.Conditions[0].LastTransitionTime = metav1.Now()
		s2.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "CreateContainerError",
						Message: "this is a transient error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s2, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseInitializing, taskStatus.Phase())
	})

	t.Run("CreateContainerErrorOutsideGracePeriod", func(t *testing.T) {
		s2 := *s.DeepCopy()
		s2.Conditions[0].LastTransitionTime.Time = metav1.Now().Add(-config.GetK8sPluginConfig().CreateContainerErrorGracePeriod.Duration)
		s2.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "CreateContainerError",
						Message: "this a permanent error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s2, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhasePermanentFailure, taskStatus.Phase())
		assert.True(t, taskStatus.CleanupOnFailure())
	})
}

func TestDemystifyPendingTimeout(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		CreateContainerErrorGracePeriod: config1.Duration{
			Duration: time.Minute * 3,
		},
		ImagePullBackoffGracePeriod: config1.Duration{
			Duration: time.Minute * 3,
		},
		PodPendingTimeout: config1.Duration{
			Duration: 10,
		},
	}))

	s := v1.PodStatus{
		Phase: v1.PodPending,
		Conditions: []v1.PodCondition{
			{
				Type:   v1.PodScheduled,
				Status: v1.ConditionFalse,
			},
		},
	}
	s.Conditions[0].LastTransitionTime.Time = metav1.Now().Add(-config.GetK8sPluginConfig().PodPendingTimeout.Duration)

	t.Run("PodPendingExceedsTimeout", func(t *testing.T) {
		taskStatus, err := DemystifyPending(s, pluginsCore.TaskInfo{})
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskStatus.Phase())
		assert.Equal(t, "PodPendingTimeout", taskStatus.Err().GetCode())
		assert.True(t, taskStatus.CleanupOnFailure())
	})
}

func TestDemystifySuccess(t *testing.T) {
	t.Run("OOMKilled", func(t *testing.T) {
		phaseInfo, err := DemystifySuccess(v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason: OOMKilled,
						},
					},
				},
			},
		}, pluginsCore.TaskInfo{})
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "OOMKilled", phaseInfo.Err().GetCode())
	})

	t.Run("InitContainer OOMKilled", func(t *testing.T) {
		phaseInfo, err := DemystifySuccess(v1.PodStatus{
			InitContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason: OOMKilled,
						},
					},
				},
			},
		}, pluginsCore.TaskInfo{})
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "OOMKilled", phaseInfo.Err().GetCode())
	})

	t.Run("success", func(t *testing.T) {
		phaseInfo, err := DemystifySuccess(v1.PodStatus{}, pluginsCore.TaskInfo{})
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseSuccess, phaseInfo.Phase())
	})
}

func TestDemystifyFailure(t *testing.T) {
	ctx := context.TODO()

	t.Run("unknown-error", func(t *testing.T) {
		phaseInfo, err := DemystifyFailure(ctx, v1.PodStatus{}, pluginsCore.TaskInfo{}, "")
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "Interrupted", phaseInfo.Err().GetCode())
		assert.Equal(t, core.ExecutionError_SYSTEM, phaseInfo.Err().GetKind())
	})

	t.Run("known-error", func(t *testing.T) {
		phaseInfo, err := DemystifyFailure(ctx, v1.PodStatus{Reason: "hello"}, pluginsCore.TaskInfo{}, "")
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "hello", phaseInfo.Err().GetCode())
		assert.Equal(t, core.ExecutionError_USER, phaseInfo.Err().GetKind())
	})

	t.Run("OOMKilled", func(t *testing.T) {
		phaseInfo, err := DemystifyFailure(ctx, v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   OOMKilled,
							ExitCode: 137,
						},
					},
				},
			},
		}, pluginsCore.TaskInfo{}, "")
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "OOMKilled", phaseInfo.Err().GetCode())
		assert.Equal(t, core.ExecutionError_USER, phaseInfo.Err().GetKind())
	})

	t.Run("SIGKILL non-primary container", func(t *testing.T) {
		phaseInfo, err := DemystifyFailure(ctx, v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					LastTerminationState: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   "some reason",
							ExitCode: SIGKILL,
						},
					},
					Name: "non-primary-container",
				},
			},
		}, pluginsCore.TaskInfo{}, "primary-container")
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "Interrupted", phaseInfo.Err().GetCode())
		assert.Equal(t, core.ExecutionError_USER, phaseInfo.Err().GetKind())
	})

	t.Run("SIGKILL primary container", func(t *testing.T) {
		phaseInfo, err := DemystifyFailure(ctx, v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					LastTerminationState: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   "some reason",
							ExitCode: SIGKILL,
						},
					},
					Name: "primary-container",
				},
			},
		}, pluginsCore.TaskInfo{}, "primary-container")
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "Interrupted", phaseInfo.Err().GetCode())
		assert.Equal(t, core.ExecutionError_SYSTEM, phaseInfo.Err().GetKind())
	})

	t.Run("GKE node preemption", func(t *testing.T) {
		for _, reason := range []string{
			"Terminated",
			"Shutdown",
			"NodeShutdown",
		} {
			t.Run(reason, func(t *testing.T) {
				message := "Test pod status message"
				phaseInfo, err := DemystifyFailure(ctx, v1.PodStatus{
					Message: message,
					Reason:  reason,
					// Can't always rely on GCP returining container statuses when node is preempted
					ContainerStatuses: []v1.ContainerStatus{},
				}, pluginsCore.TaskInfo{}, "")
				assert.Nil(t, err)
				assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
				assert.Equal(t, "Interrupted", phaseInfo.Err().GetCode())
				assert.Equal(t, core.ExecutionError_SYSTEM, phaseInfo.Err().GetKind())
				assert.Equal(t, message, phaseInfo.Err().GetMessage())
			})
		}
	})

	t.Run("Kubelet admission denies pod due to missing node label", func(t *testing.T) {
		for _, reason := range []string{
			"NodeAffinity",
		} {
			t.Run(reason, func(t *testing.T) {
				message := "Pod was rejected: Predicate NodeAffinity failed: node(s) didn't match Pod's node affinity/selector"
				phaseInfo, err := DemystifyFailure(ctx, v1.PodStatus{
					Message: message,
					Reason:  reason,
					Phase:   v1.PodFailed,
					// Can't always rely on GCP returining container statuses when node is preempted
					ContainerStatuses: []v1.ContainerStatus{},
				}, pluginsCore.TaskInfo{}, "")
				assert.Nil(t, err)
				assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
				assert.Equal(t, "Interrupted", phaseInfo.Err().GetCode())
				assert.Equal(t, core.ExecutionError_SYSTEM, phaseInfo.Err().GetKind())
				assert.Equal(t, message, phaseInfo.Err().GetMessage())
			})
		}
	})
}

func TestDemystifyPending_testcases(t *testing.T) {

	tests := []struct {
		name     string
		filename string
		isErr    bool
		errCode  string
		message  string
	}{
		{"ImagePullBackOff", "imagepull-failurepod.json", false, "ContainersNotReady|ImagePullBackOff", "Grace period [3m0s] exceeded|containers with unready status: [fdf98e4ed2b524dc3bf7-get-flyte-id-task-0]|Back-off pulling image \"image\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join("testdata", tt.filename)
			data, err := ioutil.ReadFile(testFile)
			assert.NoError(t, err, "failed to read file %s", testFile)
			pod := &v1.Pod{}
			if assert.NoError(t, json.Unmarshal(data, pod), "failed to unmarshal json in %s. Expected of type v1.Pod", testFile) {
				p, err := DemystifyPending(pod.Status, pluginsCore.TaskInfo{})
				if tt.isErr {
					assert.Error(t, err, "Error expected from method")
				} else {
					assert.NoError(t, err, "Error not expected")
					assert.NotNil(t, p)
					assert.Equal(t, p.Phase(), pluginsCore.PhaseRetryableFailure)
					if assert.NotNil(t, p.Err()) {
						assert.Equal(t, p.Err().GetCode(), tt.errCode)
						assert.Equal(t, p.Err().GetMessage(), tt.message)
					}
				}
			}
		})
	}
}

func TestDeterminePrimaryContainerPhase(t *testing.T) {
	ctx := context.TODO()
	primaryContainerName := "primary"
	secondaryContainer := v1.ContainerStatus{
		Name: "secondary",
		State: v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{
				ExitCode: 0,
			},
		},
	}
	var info = &pluginsCore.TaskInfo{}
	t.Run("primary container waiting", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(ctx, primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason: "just dawdling",
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseRunning, phaseInfo.Phase())
	})
	t.Run("primary container running", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(ctx, primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Running: &v1.ContainerStateRunning{
						StartedAt: metav1.Now(),
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseRunning, phaseInfo.Phase())
	})
	t.Run("primary container failed", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(ctx, primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 1,
						Reason:   "foo",
						Message:  "foo failed",
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "foo", phaseInfo.Err().GetCode())
		assert.Equal(t, core.ExecutionError_USER, phaseInfo.Err().GetKind())
		assert.Equal(t, "\r\n[primary] terminated with exit code (1). Reason [foo]. Message: \nfoo failed.", phaseInfo.Err().GetMessage())
	})
	t.Run("primary container failed - SIGKILL", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(ctx, primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 137,
						Reason:   "foo",
						Message:  "foo failed",
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "foo", phaseInfo.Err().GetCode())
		assert.Equal(t, core.ExecutionError_SYSTEM, phaseInfo.Err().GetKind())
		assert.Equal(t, "\r\n[primary] terminated with exit code (137). Reason [foo]. Message: \nfoo failed.", phaseInfo.Err().GetMessage())
	})
	t.Run("primary container failed - SIGKILL unsigned", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(ctx, primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 247,
						Reason:   "foo",
						Message:  "foo failed",
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "foo", phaseInfo.Err().GetCode())
		assert.Equal(t, core.ExecutionError_SYSTEM, phaseInfo.Err().GetKind())
		assert.Equal(t, "\r\n[primary] terminated with exit code (247). Reason [foo]. Message: \nfoo failed.", phaseInfo.Err().GetMessage())
	})
	t.Run("primary container succeeded", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(ctx, primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 0,
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseSuccess, phaseInfo.Phase())
	})
	t.Run("missing primary container", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(ctx, primaryContainerName, []v1.ContainerStatus{
			secondaryContainer,
		}, info)
		assert.Equal(t, pluginsCore.PhasePermanentFailure, phaseInfo.Phase())
		assert.Equal(t, PrimaryContainerNotFound, phaseInfo.Err().GetCode())
		assert.Equal(t, "Primary container [primary] not found in pod's container statuses", phaseInfo.Err().GetMessage())
	})
	t.Run("primary container failed with OOMKilled", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(ctx, primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 0,
						Reason:   OOMKilled,
						Message:  "foo failed",
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, core.ExecutionError_USER, phaseInfo.Err().GetKind())
		assert.Equal(t, OOMKilled, phaseInfo.Err().GetCode())
		assert.Equal(t, "\r\n[primary] terminated with exit code (0). Reason [OOMKilled]. Message: \nfoo failed.", phaseInfo.Err().GetMessage())
	})
	t.Run("primary container failed with OOMKilled - SIGKILL", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(ctx, primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 137,
						Reason:   OOMKilled,
						Message:  "foo failed",
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, core.ExecutionError_USER, phaseInfo.Err().GetKind())
		assert.Equal(t, OOMKilled, phaseInfo.Err().GetCode())
		assert.Equal(t, "\r\n[primary] terminated with exit code (137). Reason [OOMKilled]. Message: \nfoo failed.", phaseInfo.Err().GetMessage())
	})
}

func TestGetPodTemplate(t *testing.T) {
	ctx := context.TODO()

	podTemplate := v1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
	}

	t.Run("PodTemplateDoesNotExist", func(t *testing.T) {
		// initialize TaskExecutionContext
		task := &core.TaskTemplate{
			Type: "test",
		}

		taskReader := &pluginsCoreMock.TaskReader{}
		taskReader.On("Read", mock.Anything).Return(task, nil)

		tCtx := &pluginsCoreMock.TaskExecutionContext{}
		tCtx.EXPECT().TaskExecutionMetadata().Return(dummyTaskExecutionMetadata(&v1.ResourceRequirements{}, nil, "", nil))
		tCtx.EXPECT().TaskReader().Return(taskReader)

		// initialize PodTemplateStore
		store := NewPodTemplateStore()
		store.SetDefaultNamespace(podTemplate.Namespace)

		// validate base PodTemplate
		basePodTemplate, err := getBasePodTemplate(ctx, tCtx, store)
		assert.Nil(t, err)
		assert.Nil(t, basePodTemplate)
	})

	t.Run("PodTemplateFromTaskTemplateNameExists", func(t *testing.T) {
		// initialize TaskExecutionContext
		task := &core.TaskTemplate{
			Metadata: &core.TaskMetadata{
				PodTemplateName: "foo",
			},
			Type: "test",
		}

		taskReader := &pluginsCoreMock.TaskReader{}
		taskReader.On("Read", mock.Anything).Return(task, nil)

		tCtx := &pluginsCoreMock.TaskExecutionContext{}
		tCtx.EXPECT().TaskExecutionMetadata().Return(dummyTaskExecutionMetadata(&v1.ResourceRequirements{}, nil, "", nil))
		tCtx.EXPECT().TaskReader().Return(taskReader)

		// initialize PodTemplateStore
		store := NewPodTemplateStore()
		store.SetDefaultNamespace(podTemplate.Namespace)
		store.Store(&podTemplate)

		// validate base PodTemplate
		basePodTemplate, err := getBasePodTemplate(ctx, tCtx, store)
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(podTemplate, *basePodTemplate))
	})

	t.Run("PodTemplateFromTaskTemplateNameDoesNotExist", func(t *testing.T) {
		// initialize TaskExecutionContext
		task := &core.TaskTemplate{
			Type: "test",
			Metadata: &core.TaskMetadata{
				PodTemplateName: "foo",
			},
		}

		taskReader := &pluginsCoreMock.TaskReader{}
		taskReader.On("Read", mock.Anything).Return(task, nil)

		tCtx := &pluginsCoreMock.TaskExecutionContext{}
		tCtx.EXPECT().TaskExecutionMetadata().Return(dummyTaskExecutionMetadata(&v1.ResourceRequirements{}, nil, "", nil))
		tCtx.EXPECT().TaskReader().Return(taskReader)

		// initialize PodTemplateStore
		store := NewPodTemplateStore()
		store.SetDefaultNamespace(podTemplate.Namespace)

		// validate base PodTemplate
		basePodTemplate, err := getBasePodTemplate(ctx, tCtx, store)
		assert.NotNil(t, err)
		assert.Nil(t, basePodTemplate)
	})

	t.Run("PodTemplateFromDefaultPodTemplate", func(t *testing.T) {
		// set default PodTemplate name configuration
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			DefaultPodTemplateName: "foo",
		}))

		// initialize TaskExecutionContext
		task := &core.TaskTemplate{
			Type: "test",
		}

		taskReader := &pluginsCoreMock.TaskReader{}
		taskReader.On("Read", mock.Anything).Return(task, nil)

		tCtx := &pluginsCoreMock.TaskExecutionContext{}
		tCtx.EXPECT().TaskExecutionMetadata().Return(dummyTaskExecutionMetadata(&v1.ResourceRequirements{}, nil, "", nil))
		tCtx.EXPECT().TaskReader().Return(taskReader)

		// initialize PodTemplateStore
		store := NewPodTemplateStore()
		store.SetDefaultNamespace(podTemplate.Namespace)
		store.Store(&podTemplate)

		// validate base PodTemplate
		basePodTemplate, err := getBasePodTemplate(ctx, tCtx, store)
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(podTemplate, *basePodTemplate))
	})
}

func TestMergeWithBasePodTemplate(t *testing.T) {
	podSpec := v1.PodSpec{
		Containers: []v1.Container{
			v1.Container{
				Name: "foo",
			},
			v1.Container{
				Name: "bar",
			},
		},
		InitContainers: []v1.Container{
			v1.Container{
				Name: "foo-init",
			},
			v1.Container{
				Name: "foo-bar",
			},
		},
	}

	objectMeta := metav1.ObjectMeta{
		Labels: map[string]string{
			"fooKey": "barValue",
		},
	}

	t.Run("BasePodTemplateDoesNotExist", func(t *testing.T) {
		task := &core.TaskTemplate{
			Type: "test",
		}

		taskReader := &pluginsCoreMock.TaskReader{}
		taskReader.On("Read", mock.Anything).Return(task, nil)

		tCtx := &pluginsCoreMock.TaskExecutionContext{}
		tCtx.EXPECT().TaskExecutionMetadata().Return(dummyTaskExecutionMetadata(&v1.ResourceRequirements{}, nil, "", nil))
		tCtx.EXPECT().TaskReader().Return(taskReader)

		resultPodSpec, resultObjectMeta, err := MergeWithBasePodTemplate(context.TODO(), tCtx, &podSpec, &objectMeta, "foo", "foo-init")
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(podSpec, *resultPodSpec))
		assert.True(t, reflect.DeepEqual(objectMeta, *resultObjectMeta))
	})

	t.Run("BasePodTemplateExists", func(t *testing.T) {
		primaryContainerTemplate := v1.Container{
			Name:                   primaryContainerTemplateName,
			TerminationMessagePath: "/dev/primary-termination-log",
		}

		primaryInitContainerTemplate := v1.Container{
			Name:                   primaryInitContainerTemplateName,
			TerminationMessagePath: "/dev/primary-init-termination-log",
		}

		podTemplate := v1.PodTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "fooTemplate",
				Namespace: "test-namespace",
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"fooKey": "bazValue",
						"barKey": "bazValue",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						primaryContainerTemplate,
					},
					InitContainers: []v1.Container{
						primaryInitContainerTemplate,
					},
				},
			},
		}

		DefaultPodTemplateStore.Store(&podTemplate)

		task := &core.TaskTemplate{
			Metadata: &core.TaskMetadata{
				PodTemplateName: "fooTemplate",
			},
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Command: []string{"command"},
					Args:    []string{"{{.Input}}"},
				},
			},
			Type: "test",
		}

		taskReader := &pluginsCoreMock.TaskReader{}
		taskReader.On("Read", mock.Anything).Return(task, nil)

		tCtx := &pluginsCoreMock.TaskExecutionContext{}
		tCtx.EXPECT().TaskExecutionMetadata().Return(dummyTaskExecutionMetadata(&v1.ResourceRequirements{}, nil, "", nil))
		tCtx.EXPECT().TaskReader().Return(taskReader)

		resultPodSpec, resultObjectMeta, err := MergeWithBasePodTemplate(context.TODO(), tCtx, &podSpec, &objectMeta, "foo", "foo-init")
		assert.Nil(t, err)

		// test that template podSpec is merged
		primaryContainer := resultPodSpec.Containers[0]
		assert.Equal(t, podSpec.Containers[0].Name, primaryContainer.Name)
		assert.Equal(t, primaryContainerTemplate.TerminationMessagePath, primaryContainer.TerminationMessagePath)
		primaryInitContainer := resultPodSpec.InitContainers[0]
		assert.Equal(t, podSpec.InitContainers[0].Name, primaryInitContainer.Name)
		assert.Equal(t, primaryInitContainerTemplate.TerminationMessagePath, primaryInitContainer.TerminationMessagePath)

		// test that template object metadata is copied
		assert.Contains(t, resultObjectMeta.Labels, "fooKey")
		assert.Equal(t, resultObjectMeta.Labels["fooKey"], "barValue")
		assert.Contains(t, resultObjectMeta.Labels, "barKey")
		assert.Equal(t, resultObjectMeta.Labels["barKey"], "bazValue")
	})
}

func TestMergeBasePodSpecsOntoTemplate(t *testing.T) {

	baseContainer1 := v1.Container{
		Name:  "task-1",
		Image: "task-image",
	}

	baseContainer2 := v1.Container{
		Name:  "task-2",
		Image: "task-image",
	}

	initContainer1 := v1.Container{
		Name:  "task-init-1",
		Image: "task-init-image",
	}

	initContainer2 := v1.Container{
		Name:  "task-init-2",
		Image: "task-init-image",
	}

	tests := []struct {
		name                     string
		templatePodSpec          *v1.PodSpec
		basePodSpec              *v1.PodSpec
		primaryContainerName     string
		primaryInitContainerName string
		expectedResult           *v1.PodSpec
		expectedError            error
	}{
		{
			name:            "nil template",
			templatePodSpec: nil,
			basePodSpec:     &v1.PodSpec{},
			expectedError:   errors.New("neither the templatePodSpec or the basePodSpec can be nil"),
		},
		{
			name:            "nil base",
			templatePodSpec: &v1.PodSpec{},
			basePodSpec:     nil,
			expectedError:   errors.New("neither the templatePodSpec or the basePodSpec can be nil"),
		},
		{
			name:            "nil template and base",
			templatePodSpec: nil,
			basePodSpec:     nil,
			expectedError:   errors.New("neither the templatePodSpec or the basePodSpec can be nil"),
		},
		{
			name: "template and base with no overlap",
			templatePodSpec: &v1.PodSpec{
				SchedulerName: "templateScheduler",
			},
			basePodSpec: &v1.PodSpec{
				ServiceAccountName: "baseServiceAccount",
			},
			expectedResult: &v1.PodSpec{
				SchedulerName:      "templateScheduler",
				ServiceAccountName: "baseServiceAccount",
			},
		},
		{
			name: "template and base with overlap",
			templatePodSpec: &v1.PodSpec{
				SchedulerName: "templateScheduler",
			},
			basePodSpec: &v1.PodSpec{
				SchedulerName:      "baseScheduler",
				ServiceAccountName: "baseServiceAccount",
			},
			expectedResult: &v1.PodSpec{
				SchedulerName:      "baseScheduler",
				ServiceAccountName: "baseServiceAccount",
			},
		},
		{
			name: "template with default containers and base with no containers",
			templatePodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "default",
						Image: "default-image",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:  "default-init",
						Image: "default-init-image",
					},
				},
			},
			basePodSpec: &v1.PodSpec{
				SchedulerName: "baseScheduler",
			},
			expectedResult: &v1.PodSpec{
				SchedulerName: "baseScheduler",
			},
		},
		{
			name:            "template with no default containers and base containers",
			templatePodSpec: &v1.PodSpec{},
			basePodSpec: &v1.PodSpec{
				Containers:     []v1.Container{baseContainer1},
				InitContainers: []v1.Container{initContainer1},
				SchedulerName:  "baseScheduler",
			},
			expectedResult: &v1.PodSpec{
				Containers:     []v1.Container{baseContainer1},
				InitContainers: []v1.Container{initContainer1},
				SchedulerName:  "baseScheduler",
			},
		},
		{
			name: "template and base with matching containers",
			templatePodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:                   "task-1",
						Image:                  "default-task-image",
						TerminationMessagePath: "/dev/template-termination-log",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:                   "task-init-1",
						Image:                  "default-task-init-image",
						TerminationMessagePath: "/dev/template-init-termination-log",
					},
				},
			},
			basePodSpec: &v1.PodSpec{
				Containers:     []v1.Container{baseContainer1},
				InitContainers: []v1.Container{initContainer1},
				SchedulerName:  "baseScheduler",
			},
			expectedResult: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:                   "task-1",
						Image:                  "task-image",
						TerminationMessagePath: "/dev/template-termination-log",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:                   "task-init-1",
						Image:                  "task-init-image",
						TerminationMessagePath: "/dev/template-init-termination-log",
					},
				},
				SchedulerName: "baseScheduler",
			},
		},
		{
			name: "template and base with no matching containers",
			templatePodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:                   "not-matching",
						Image:                  "default-task-image",
						TerminationMessagePath: "/dev/template-termination-log",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:                   "not-matching-init",
						Image:                  "default-task-init-image",
						TerminationMessagePath: "/dev/template-init-termination-log",
					},
				},
			},
			basePodSpec: &v1.PodSpec{
				Containers:     []v1.Container{baseContainer1},
				InitContainers: []v1.Container{initContainer1},
				SchedulerName:  "baseScheduler",
			},
			expectedResult: &v1.PodSpec{
				Containers:     []v1.Container{baseContainer1},
				InitContainers: []v1.Container{initContainer1},
				SchedulerName:  "baseScheduler",
			},
		},
		{
			name: "template with default containers and base with containers",
			templatePodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:                   "default",
						Image:                  "default-task-image",
						TerminationMessagePath: "/dev/template-termination-log",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:                   "default-init",
						Image:                  "default-task-init-image",
						TerminationMessagePath: "/dev/template-init-termination-log",
					},
				},
			},
			basePodSpec: &v1.PodSpec{
				Containers:     []v1.Container{baseContainer1, baseContainer2},
				InitContainers: []v1.Container{initContainer1, initContainer2},
				SchedulerName:  "baseScheduler",
			},
			expectedResult: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:                   "task-1",
						Image:                  "task-image",
						TerminationMessagePath: "/dev/template-termination-log",
					},
					{
						Name:                   "task-2",
						Image:                  "task-image",
						TerminationMessagePath: "/dev/template-termination-log",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:                   "task-init-1",
						Image:                  "task-init-image",
						TerminationMessagePath: "/dev/template-init-termination-log",
					},
					{
						Name:                   "task-init-2",
						Image:                  "task-init-image",
						TerminationMessagePath: "/dev/template-init-termination-log",
					},
				},
				SchedulerName: "baseScheduler",
			},
		},
		{
			name: "template with primary containers and base with containers",
			templatePodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:                   "primary",
						Image:                  "default-task-image",
						TerminationMessagePath: "/dev/template-termination-log",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:                   "primary-init",
						Image:                  "default-task-init-image",
						TerminationMessagePath: "/dev/template-init-termination-log",
					},
				},
			},
			basePodSpec: &v1.PodSpec{
				Containers:     []v1.Container{baseContainer1, baseContainer2},
				InitContainers: []v1.Container{initContainer1, initContainer2},
				SchedulerName:  "baseScheduler",
			},
			expectedResult: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:                   "task-1",
						Image:                  "task-image",
						TerminationMessagePath: "/dev/template-termination-log",
					},
					baseContainer2,
				},
				InitContainers: []v1.Container{
					{
						Name:                   "task-init-1",
						Image:                  "task-init-image",
						TerminationMessagePath: "/dev/template-init-termination-log",
					},
					initContainer2,
				},
				SchedulerName: "baseScheduler",
			},
			primaryContainerName:     "task-1",
			primaryInitContainerName: "task-init-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, mergeErr := MergeBasePodSpecOntoTemplate(tt.templatePodSpec, tt.basePodSpec, tt.primaryContainerName, tt.primaryInitContainerName)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedError, mergeErr)
		})
	}
}

func TestMergeOverlayPodSpecOntoBase(t *testing.T) {

	tests := []struct {
		name           string
		basePodSpec    *v1.PodSpec
		overlayPodSpec *v1.PodSpec
		expectedResult *v1.PodSpec
		expectedError  error
	}{
		{
			name:           "nil overlay",
			basePodSpec:    &v1.PodSpec{},
			overlayPodSpec: nil,
			expectedError:  errors.New("neither the basePodSpec or the overlayPodSpec can be nil"),
		},
		{
			name:           "nil base",
			basePodSpec:    nil,
			overlayPodSpec: &v1.PodSpec{},
			expectedError:  errors.New("neither the basePodSpec or the overlayPodSpec can be nil"),
		},
		{
			name:           "nil base and overlay",
			basePodSpec:    nil,
			overlayPodSpec: nil,
			expectedError:  errors.New("neither the basePodSpec or the overlayPodSpec can be nil"),
		},
		{
			name: "base and overlay no overlap",
			basePodSpec: &v1.PodSpec{
				SchedulerName: "baseScheduler",
			},
			overlayPodSpec: &v1.PodSpec{
				ServiceAccountName: "overlayServiceAccount",
			},
			expectedResult: &v1.PodSpec{
				SchedulerName:      "baseScheduler",
				ServiceAccountName: "overlayServiceAccount",
			},
		},
		{
			name: "template and base with overlap",
			basePodSpec: &v1.PodSpec{
				SchedulerName: "baseScheduler",
			},
			overlayPodSpec: &v1.PodSpec{
				SchedulerName:      "overlayScheduler",
				ServiceAccountName: "overlayServiceAccount",
			},
			expectedResult: &v1.PodSpec{
				SchedulerName:      "overlayScheduler",
				ServiceAccountName: "overlayServiceAccount",
			},
		},
		{
			name: "template and base with matching containers",
			basePodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "task-1",
						Image: "task-image",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:  "task-init-1",
						Image: "task-init-image",
					},
				},
			},
			overlayPodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "task-1",
						Image: "overlay-image",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:  "task-init-1",
						Image: "overlay-init-image",
					},
				},
			},
			expectedResult: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "task-1",
						Image: "overlay-image",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:  "task-init-1",
						Image: "overlay-init-image",
					},
				},
			},
		},
		{
			name: "base and overlay with no matching containers",
			basePodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "task-1",
						Image: "task-image",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:  "task-init-1",
						Image: "task-init-image",
					},
				},
			},
			overlayPodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "overlay-1",
						Image: "overlay-image",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:  "overlay-init-1",
						Image: "overlay-init-image",
					},
				},
			},
			expectedResult: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "task-1",
						Image: "task-image",
					},
				},
				InitContainers: []v1.Container{
					{
						Name:  "task-init-1",
						Image: "task-init-image",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, mergeErr := MergeOverlayPodSpecOntoBase(tt.basePodSpec, tt.overlayPodSpec)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedError, mergeErr)
		})
	}
}

func TestAddFlyteCustomizationsToContainer_SetConsoleUrl(t *testing.T) {
	tests := []struct {
		name              string
		includeConsoleURL bool
		consoleURL        string
		expectedEnvVar    *v1.EnvVar
	}{
		{
			name:              "do not include console url and console url is not set",
			includeConsoleURL: false,
			consoleURL:        "",
			expectedEnvVar:    nil,
		},
		{
			name:              "include console url but console url is not set",
			includeConsoleURL: false,
			consoleURL:        "",
			expectedEnvVar:    nil,
		},
		{
			name:              "do not include console url but console url is set",
			includeConsoleURL: false,
			consoleURL:        "gopher://flyte:65535/console",
			expectedEnvVar:    nil,
		},
		{
			name:              "include console url and console url is set",
			includeConsoleURL: true,
			consoleURL:        "gopher://flyte:65535/console",
			expectedEnvVar: &v1.EnvVar{
				Name:  flyteExecutionURL,
				Value: "gopher://flyte:65535/console/projects/p2/domains/d2/executions/n2/nodeId/unique_node_id/nodes",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := &v1.Container{
				Command: []string{
					"{{ .Input }}",
				},
				Args: []string{
					"{{ .OutputPrefix }}",
				},
			}
			templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{}, &v1.ResourceRequirements{}, tt.includeConsoleURL, tt.consoleURL)
			err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeAssignResources, container)
			assert.NoError(t, err)
			if tt.expectedEnvVar == nil {
				// Confirm that there is no env var FLYTE_EXECUTION_URL set
				for _, envVar := range container.Env {
					assert.NotEqual(t, "FLYTE_EXECUTION_URL", envVar.Name)
				}
			}
			if tt.expectedEnvVar != nil {
				// Assert that the env var FLYTE_EXECUTION_URL is set if its value is non-nil
				for _, envVar := range container.Env {
					if envVar.Name == tt.expectedEnvVar.Name {
						assert.Equal(t, tt.expectedEnvVar.Value, envVar.Value)
						return
					}
				}
				t.Fail()
			}
		})
	}
}

func TestAddTolerationsForExtendedResources(t *testing.T) {
	gpuResourceName := v1.ResourceName("nvidia.com/gpu")
	addTolerationResourceName := v1.ResourceName("foo/bar")
	noTolerationResourceName := v1.ResourceName("foo/baz")
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		GpuResourceName: gpuResourceName,
		AddTolerationsForExtendedResources: []string{
			gpuResourceName.String(),
			addTolerationResourceName.String(),
		},
	}))

	podSpec := &v1.PodSpec{
		Containers: []v1.Container{
			v1.Container{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						gpuResourceName:           resource.MustParse("1"),
						addTolerationResourceName: resource.MustParse("1"),
						noTolerationResourceName:  resource.MustParse("1"),
					},
				},
			},
		},
		Tolerations: []v1.Toleration{
			{
				Key:      "foo",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
	}

	podSpec = AddTolerationsForExtendedResources(podSpec)
	fmt.Printf("%v\n", podSpec.Tolerations)
	assert.Equal(t, 3, len(podSpec.Tolerations))
	assert.Equal(t, addTolerationResourceName.String(), podSpec.Tolerations[1].Key)
	assert.Equal(t, v1.TolerationOpExists, podSpec.Tolerations[1].Operator)
	assert.Equal(t, v1.TaintEffectNoSchedule, podSpec.Tolerations[1].Effect)
	assert.Equal(t, gpuResourceName.String(), podSpec.Tolerations[2].Key)
	assert.Equal(t, v1.TolerationOpExists, podSpec.Tolerations[2].Operator)
	assert.Equal(t, v1.TaintEffectNoSchedule, podSpec.Tolerations[2].Effect)

	podSpec = &v1.PodSpec{
		InitContainers: []v1.Container{
			v1.Container{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						gpuResourceName:           resource.MustParse("1"),
						addTolerationResourceName: resource.MustParse("1"),
						noTolerationResourceName:  resource.MustParse("1"),
					},
				},
			},
		},
		Tolerations: []v1.Toleration{
			{
				Key:      "foo",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
	}

	podSpec = AddTolerationsForExtendedResources(podSpec)
	assert.Equal(t, 3, len(podSpec.Tolerations))
	assert.Equal(t, addTolerationResourceName.String(), podSpec.Tolerations[1].Key)
	assert.Equal(t, v1.TolerationOpExists, podSpec.Tolerations[1].Operator)
	assert.Equal(t, v1.TaintEffectNoSchedule, podSpec.Tolerations[1].Effect)
	assert.Equal(t, gpuResourceName.String(), podSpec.Tolerations[2].Key)
	assert.Equal(t, v1.TolerationOpExists, podSpec.Tolerations[2].Operator)
	assert.Equal(t, v1.TaintEffectNoSchedule, podSpec.Tolerations[2].Effect)

	podSpec = &v1.PodSpec{
		Containers: []v1.Container{
			v1.Container{
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						gpuResourceName:           resource.MustParse("1"),
						addTolerationResourceName: resource.MustParse("1"),
						noTolerationResourceName:  resource.MustParse("1"),
					},
				},
			},
		},
		Tolerations: []v1.Toleration{
			{
				Key:      "foo",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
			{
				Key:      gpuResourceName.String(),
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
	}

	podSpec = AddTolerationsForExtendedResources(podSpec)
	assert.Equal(t, 3, len(podSpec.Tolerations))
	assert.Equal(t, gpuResourceName.String(), podSpec.Tolerations[1].Key)
	assert.Equal(t, v1.TolerationOpExists, podSpec.Tolerations[1].Operator)
	assert.Equal(t, v1.TaintEffectNoSchedule, podSpec.Tolerations[1].Effect)
	assert.Equal(t, addTolerationResourceName.String(), podSpec.Tolerations[2].Key)
	assert.Equal(t, v1.TolerationOpExists, podSpec.Tolerations[2].Operator)
	assert.Equal(t, v1.TaintEffectNoSchedule, podSpec.Tolerations[2].Effect)
}

func TestApplyExtendedResourcesOverridesSharedMemory(t *testing.T) {
	SharedMemory := &core.ExtendedResources{
		SharedMemory: &core.SharedMemory{
			MountName: "flyte-shared-memory",
			MountPath: "/dev/shm",
		},
	}

	newSharedMemory := &core.ExtendedResources{
		SharedMemory: &core.SharedMemory{
			MountName: "flyte-shared-memory-v2",
			MountPath: "/dev/shm",
		},
	}

	t.Run("base is nil", func(t *testing.T) {
		final := applyExtendedResourcesOverrides(nil, SharedMemory)
		assert.EqualValues(
			t,
			SharedMemory.GetSharedMemory(),
			final.GetSharedMemory(),
		)
	})

	t.Run("overrides is nil", func(t *testing.T) {
		final := applyExtendedResourcesOverrides(SharedMemory, nil)
		assert.EqualValues(
			t,
			SharedMemory.GetSharedMemory(),
			final.GetSharedMemory(),
		)
	})

	t.Run("merging", func(t *testing.T) {
		final := applyExtendedResourcesOverrides(SharedMemory, newSharedMemory)
		assert.EqualValues(
			t,
			newSharedMemory.GetSharedMemory(),
			final.GetSharedMemory(),
		)
	})
}

func TestApplySharedMemoryErrors(t *testing.T) {

	type test struct {
		name                 string
		podSpec              *v1.PodSpec
		primaryContainerName string
		sharedVolume         *core.SharedMemory
		errorMsg             string
	}

	tests := []test{
		{
			name:                 "No mount name",
			podSpec:              nil,
			primaryContainerName: "primary",
			sharedVolume:         &core.SharedMemory{MountPath: "/dev/shm"},
			errorMsg:             "mount name is not set",
		},
		{
			name:                 "No mount path name",
			podSpec:              nil,
			primaryContainerName: "primary",
			sharedVolume:         &core.SharedMemory{MountName: "flyte-shared-memory"},
			errorMsg:             "mount path is not set",
		},
		{
			name: "No primary container",
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name: "secondary",
				}},
			},
			primaryContainerName: "primary",
			sharedVolume:         &core.SharedMemory{MountName: "flyte-shared-memory", MountPath: "/dev/shm"},
			errorMsg:             "Unable to find primary container",
		},

		{
			name: "Volume already exists in spec",
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name: "primary",
				}},
				Volumes: []v1.Volume{{
					Name: "flyte-shared-memory",
				}},
			},
			primaryContainerName: "primary",
			sharedVolume:         &core.SharedMemory{MountName: "flyte-shared-memory", MountPath: "/dev/shm"},
			errorMsg:             "A volume is already named flyte-shared-memory in pod spec",
		},
		{
			name: "Volume already in container",
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name: "primary",
					VolumeMounts: []v1.VolumeMount{{
						Name:      "flyte-shared-memory",
						MountPath: "/dev/shm",
					}},
				}},
			},
			primaryContainerName: "primary",
			sharedVolume:         &core.SharedMemory{MountName: "flyte-shared-memory", MountPath: "/dev/shm"},
			errorMsg:             "A volume is already named flyte-shared-memory in container",
		},
		{
			name: "Mount path already in container",
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name: "primary",
					VolumeMounts: []v1.VolumeMount{{
						Name:      "flyte-shared-memory-v2",
						MountPath: "/dev/shm",
					}},
				}},
			},
			primaryContainerName: "primary",
			sharedVolume:         &core.SharedMemory{MountName: "flyte-shared-memory", MountPath: "/dev/shm"},
			errorMsg:             "/dev/shm is already mounted in container",
		},
		{
			name: "Mount path already in container",
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name: "primary",
				}},
			},
			primaryContainerName: "primary",
			sharedVolume:         &core.SharedMemory{MountName: "flyte-shared-memory", MountPath: "/dev/shm", SizeLimit: "bad-name"},
			errorMsg:             "Unable to parse size limit: bad-name",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ApplySharedMemory(test.podSpec, test.primaryContainerName, test.sharedVolume)
			assert.Errorf(t, err, test.errorMsg)
		})
	}
}

func TestApplySharedMemory(t *testing.T) {

	type test struct {
		name                 string
		podSpec              *v1.PodSpec
		primaryContainerName string
		sharedVolume         *core.SharedMemory
	}

	tests := []test{
		{
			name: "No size limit works",
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name: "primary",
				}},
			},
			primaryContainerName: "primary",
			sharedVolume:         &core.SharedMemory{MountName: "flyte-shared-memory", MountPath: "/dev/shm"},
		},
		{
			name: "With size limits works",
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{{
					Name: "primary",
				}},
			},
			primaryContainerName: "primary",
			sharedVolume:         &core.SharedMemory{MountName: "flyte-shared-memory", MountPath: "/dev/shm", SizeLimit: "2Gi"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ApplySharedMemory(test.podSpec, test.primaryContainerName, test.sharedVolume)
			assert.NoError(t, err)

			assert.Len(t, test.podSpec.Volumes, 1)
			assert.Len(t, test.podSpec.Containers[0].VolumeMounts, 1)

			assert.Equal(
				t,
				test.podSpec.Containers[0].VolumeMounts[0],
				v1.VolumeMount{
					Name:      test.sharedVolume.GetMountName(),
					MountPath: test.sharedVolume.GetMountPath(),
				},
			)

			var quantity resource.Quantity
			if test.sharedVolume.GetSizeLimit() != "" {
				quantity, err = resource.ParseQuantity(test.sharedVolume.GetSizeLimit())
				assert.NoError(t, err)
			}

			assert.Equal(
				t,
				test.podSpec.Volumes[0],
				v1.Volume{
					Name: test.sharedVolume.GetMountName(),
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{Medium: v1.StorageMediumMemory, SizeLimit: &quantity},
					},
				},
			)

		})
	}
}
