package coach

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type usageMetrics struct {
	Stage            string   `json:"stage"`
	Model            string   `json:"model"`
	Outcome          string   `json:"outcome"`
	ElapsedMs        int64    `json:"elapsedMs"`
	InputTokens      *int     `json:"inputTokens"`
	OutputTokens     *int     `json:"outputTokens"`
	EstimatedCostUSD *float64 `json:"estimatedCostUSD"`
}

// Prices are explicitly configured for the deployed model, never inferred from
// a demo or a stale price table. Missing usage or prices remain unknown (null).
func providerMetrics(stage, model, outcome string, elapsed time.Duration, raw json.RawMessage) usageMetrics {
	m := usageMetrics{Stage: stage, Model: model, Outcome: outcome, ElapsedMs: elapsed.Milliseconds()}
	var usage struct {
		Input  *int `json:"input_tokens"`
		Output *int `json:"output_tokens"`
	}
	if json.Unmarshal(raw, &usage) != nil || usage.Input == nil || usage.Output == nil || *usage.Input < 0 || *usage.Output < 0 {
		return m
	}
	m.InputTokens = usage.Input
	m.OutputTokens = usage.Output
	prefix := "COACH_" + strings.ToUpper(stage)
	inputRate, e1 := strconv.ParseFloat(os.Getenv(prefix+"_INPUT_USD_PER_MILLION"), 64)
	outputRate, e2 := strconv.ParseFloat(os.Getenv(prefix+"_OUTPUT_USD_PER_MILLION"), 64)
	if e1 != nil || e2 != nil || math.IsNaN(inputRate) || math.IsInf(inputRate, 0) || math.IsNaN(outputRate) || math.IsInf(outputRate, 0) || inputRate < 0 || outputRate < 0 {
		return m
	}
	cost := (float64(*usage.Input)*inputRate + float64(*usage.Output)*outputRate) / 1e6
	if math.IsInf(cost, 0) || math.IsNaN(cost) {
		return m
	}
	m.EstimatedCostUSD = &cost
	return m
}
func logProvider(stage, model, outcome string, start time.Time, usage json.RawMessage) {
	data, _ := json.Marshal(providerMetrics(stage, model, outcome, time.Since(start), usage))
	log.Printf("coach provider=%s", data)
}
