package adaptivetopk

import "testing"

func TestConfigProcessorType(t *testing.T) {
	cfg := &Config{}
	if got := cfg.ProcessorType(); got != "adaptivetopk" {
		t.Errorf("expected adaptivetopk, got %s", got)
	}
}
