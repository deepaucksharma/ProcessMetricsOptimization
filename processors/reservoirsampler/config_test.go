package reservoirsampler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigProcessorType(t *testing.T) {
	cfg := &Config{}
	assert.Equal(t, "reservoirsampler", cfg.ProcessorType())
}
