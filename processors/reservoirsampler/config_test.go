package reservoirsampler

import "testing"

func TestConfigProcessorType(t *testing.T) {
	cfg := &Config{}
	if got := cfg.ProcessorType(); got != "reservoirsampler" {
		t.Errorf("expected reservoirsampler, got %s", got)
	}
}
