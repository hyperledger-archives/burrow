package config

import (
	"testing"

	"github.com/eris-ltd/eris-db/logging/config/types"
	. "github.com/eris-ltd/eris-db/util/slice"
	"github.com/stretchr/testify/assert"
)

func TestBuildLoggerFromSinkConfig(t *testing.T) {
	sinkConfig := &types.SinkConfig{
		Transform: &types.TransformConfig{
			TransformType: types.NoTransform,
		},
		Sinks: []*types.SinkConfig{
			{
				Transform: &types.TransformConfig{
					TransformType: types.NoTransform,
				},
				Sinks: []*types.SinkConfig{
					{
						Transform: &types.TransformConfig{
							TransformType: types.Capture,
							CaptureConfig: &types.CaptureConfig{
								Name:      "cap",
								BufferCap: 100,
							},
						},
						Output: &types.OutputConfig{
							OutputType: types.Stderr,
						},
						Sinks: []*types.SinkConfig{
							{
								Transform: &types.TransformConfig{
									TransformType: types.Label,
									LabelConfig: &types.LabelConfig{
										Prefix: true,
										Labels: map[string]string{"Label": "A Label!"},
									},
								},
								Output: &types.OutputConfig{
									OutputType: types.Stdout,
								},
							},
						},
					},
				},
			},
		},
	}
	logger, captures, err := BuildLoggerFromRootSinkConfig(sinkConfig)
	logger.Log("Foo", "Bar")
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"Foo", "Bar"}, captures["cap"].WaitReadLogLine())
}
