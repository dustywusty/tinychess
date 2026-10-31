package coach

import "fmt"

const EngineRevision = "361840f407b588252a48b25bb3000c751a470c89"

// Metadata describes a specific search, not an assertion of chess certainty.
type AnalysisMetadata struct {
	ID               string `json:"id"`
	Engine           string `json:"engine"`
	EngineRevision   string `json:"engineRevision"`
	RootPositionID   string `json:"rootPositionId"`
	RootSide         string `json:"rootSide"`
	ScorePerspective string `json:"scorePerspective"`
	SearchLimitMs    int    `json:"searchLimitMs"`
	SearchElapsedMs  int64  `json:"searchElapsedMs"`
	Depth            int    `json:"depth"`
	Status           string `json:"status"`
	Threads          int    `json:"threads"`
	HashMB           int    `json:"hashMB"`
}

func quality(line Line) string {
	if line.Analysis.ID == "" {
		return "Bounded engine search · provisional result"
	}
	return fmt.Sprintf("%s · depth %d · %d ms search (limit %d ms) · provisional result", line.Analysis.Engine, line.Depth, line.Analysis.SearchElapsedMs, line.Analysis.SearchLimitMs)
}
