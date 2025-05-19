package processorhelper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
)

func TestNewHelper(t *testing.T) {
	h, err := NewHelper(component.TelemetrySettings{}, "test")
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, h.Obsreport)
	require.Equal(t, true, h.Capabilities().MutatesData)
}
