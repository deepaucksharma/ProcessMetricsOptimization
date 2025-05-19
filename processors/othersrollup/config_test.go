package othersrollup

import "testing"

func TestConfigProcessorType(t *testing.T) {
	cfg := &Config{}
	if got := cfg.ProcessorType(); got != "othersrollup" {
		t.Errorf("expected othersrollup, got %s", got)
	}
}
