package coach

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTelemetryDoesNotInventMissingUsageOrPrices(t *testing.T) {
	t.Setenv("COACH_JEV_INPUT_USD_PER_MILLION", "")
	t.Setenv("COACH_JEV_OUTPUT_USD_PER_MILLION", "")
	raw := json.RawMessage(`{"input_tokens":1000,"output_tokens":200}`)
	m := providerMetrics("jev", "configured-model", "ok", time.Second, raw)
	if m.EstimatedCostUSD != nil || m.InputTokens == nil || *m.InputTokens != 1000 {
		t.Fatal(m)
	}
	t.Setenv("COACH_JEV_INPUT_USD_PER_MILLION", "2")
	t.Setenv("COACH_JEV_OUTPUT_USD_PER_MILLION", "5")
	m = providerMetrics("jev", "configured-model", "ok", time.Second, raw)
	if m.EstimatedCostUSD == nil || *m.EstimatedCostUSD != 0.003 {
		t.Fatal(m)
	}
	m = providerMetrics("jev", "configured-model", "error", time.Second, nil)
	if m.EstimatedCostUSD != nil || m.InputTokens != nil {
		t.Fatal("failed call reported invented zero cost")
	}
}
