// Package coach owns temporary analysis. It never writes to a live game.
package coach

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/corentings/chess/v2"
	"github.com/google/uuid"
)

type Score struct {
	Kind  string `json:"kind"`
	Value int    `json:"value"`
}
type Line struct {
	Analysis AnalysisMetadata `json:"analysis"`
	Moves    []string         `json:"moves"`
	SAN      []string         `json:"san"`
	FENs     []string         `json:"fens"`
	Score    Score            `json:"score"`
	Depth    int              `json:"depth"`
	Draws    map[int]string   `json:"-"`
}
type Analyzer interface {
	Analyze(context.Context, []string, string) (Line, error)
}
type Engine struct {
	Path  string
	slots chan struct{}
}

func NewEngine(path string) *Engine { return &Engine{Path: path, slots: make(chan struct{}, 2)} }

// The pinned rules version uses a shared scratch buffer while parsing FEN.
// Initialize the immutable starting game and its legal-move cache once, then
// clone it so concurrent coach searches never re-enter that parser.
var startingGame = func() *chess.Game {
	g := chess.NewGame()
	_ = g.ValidMoves()
	return g
}()

func Replay(moves []string) (*chess.Game, error) {
	g := startingGame.Clone()
	for _, u := range moves {
		if err := play(g, u); err != nil {
			return nil, err
		}
	}
	return g, nil
}
func play(g *chess.Game, u string) error {
	if g.Outcome() != chess.NoOutcome {
		return errors.New("the position is terminal")
	}
	m, err := (chess.UCINotation{}).Decode(g.Position(), u)
	if err != nil {
		return err
	}
	for _, legal := range g.ValidMoves() {
		if legal.S1() == m.S1() && legal.S2() == m.S2() && legal.Promo() == m.Promo() {
			return g.Move(&legal, nil)
		}
	}
	return errors.New("illegal move")
}

// ValidateLine derives notation and positions from rules, never engine prose.
func ValidateLine(history []string, raw Line) (Line, error) {
	g, err := Replay(history)
	if err != nil {
		return Line{}, err
	}
	if len(raw.Moves) == 0 || raw.Depth < 1 || (raw.Score.Kind != "cp" && raw.Score.Kind != "mate") {
		return Line{}, errors.New("incomplete engine analysis")
	}
	raw.FENs = []string{g.Position().String()}
	raw.SAN = nil
	raw.Draws = nil
	if len(raw.Moves) > 12 {
		raw.Moves = raw.Moves[:12]
	}
	for _, u := range raw.Moves {
		m, err := (chess.UCINotation{}).Decode(g.Position(), u)
		if err != nil {
			return Line{}, err
		}
		san := (chess.AlgebraicNotation{}).Encode(g.Position(), m)
		if err = play(g, u); err != nil {
			return Line{}, err
		}
		raw.SAN = append(raw.SAN, san)
		raw.FENs = append(raw.FENs, g.Position().String())
		if g.Outcome() != chess.NoOutcome {
			if g.Outcome() == chess.Draw {
				raw.Draws = map[int]string{len(raw.SAN): g.Method().String()}
			}
			raw.Moves = raw.Moves[:len(raw.SAN)]
			break
		}
	}
	return raw, nil
}
func (e *Engine) Analyze(ctx context.Context, history []string, candidate string) (Line, error) {
	if e.Path == "" {
		return Line{}, errors.New("engine unavailable")
	}
	g, err := Replay(history)
	if err != nil {
		return Line{}, err
	}
	rootSide := g.Position().Turn().String()
	if g.Outcome() != chess.NoOutcome {
		return Line{}, errors.New("the position is terminal; review an earlier move")
	}
	if candidate != "" {
		if err = play(g, candidate); err != nil {
			return Line{}, err
		}
	}
	select {
	case e.slots <- struct{}{}:
		defer func() { <-e.slots }()
	default:
		return Line{}, errors.New("engine busy; retry")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, e.Path)
	cmd.Dir = filepath.Dir(e.Path)
	in, err := cmd.StdinPipe()
	if err != nil {
		return Line{}, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return Line{}, err
	}
	if err = cmd.Start(); err != nil {
		return Line{}, err
	}
	defer func() { _ = in.Close(); _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	scanner := bufio.NewScanner(out)
	send := func(s string) { _, _ = fmt.Fprintln(in, s) }
	wait := func(target string) bool {
		for scanner.Scan() {
			if scanner.Text() == target {
				return true
			}
		}
		return false
	}
	send("uci")
	if !wait("uciok") {
		return Line{}, errors.New("engine handshake failed")
	}
	for _, s := range []string{"Threads value 1", "Hash value 16", "OwnBook value false", "Ponder value false", "MultiPV value 1"} {
		send("setoption name " + s)
	}
	send("isready")
	if !wait("readyok") {
		return Line{}, errors.New("engine unavailable")
	}
	send("position startpos moves " + strings.Join(history, " "))
	// Infinite analysis avoids Arasan's depth-2 early exit for a single
	// restricted root move. Stop it explicitly at the same fixed budget.
	search := "go infinite"
	if candidate != "" {
		search += " searchmoves " + candidate
	}
	searchStarted := time.Now()
	send(search)
	stopSearch := time.AfterFunc(1500*time.Millisecond, func() { send("stop") })
	defer stopSearch.Stop()
	var result Line
	for scanner.Scan() {
		s := scanner.Text()
		if strings.HasPrefix(s, "bestmove ") {
			if candidate != "" && (len(result.Moves) == 0 || result.Moves[0] != candidate) {
				return Line{}, errors.New("candidate mismatch")
			}
			best := strings.Fields(s)
			if len(best) < 2 || len(result.Moves) == 0 || best[1] != result.Moves[0] {
				return Line{}, errors.New("engine result has no matching exact-score continuation")
			}
			result.Analysis = AnalysisMetadata{ID: uuid.NewString(), Engine: "Arasan 26.0", EngineRevision: EngineRevision, RootPositionID: Identity(history), RootSide: rootSide, ScorePerspective: "root-side-to-move", SearchLimitMs: 1500, SearchElapsedMs: time.Since(searchStarted).Milliseconds(), Depth: result.Depth, Status: "complete", Threads: 1, HashMB: 16}
			return ValidateLine(history, result)
		}
		if line, ok := parseInfo(s); ok && line.Depth >= result.Depth {
			result = line
		}
	}
	return Line{}, errors.New("engine timed out or stopped")
}
func parseInfo(s string) (Line, bool) {
	f := strings.Fields(s)
	var l Line
	score := false
	if len(f) == 0 || f[0] != "info" {
		return l, false
	}
	for i := 1; i < len(f); i++ {
		switch f[i] {
		case "lowerbound", "upperbound":
			return Line{}, false
		case "depth":
			if i+1 < len(f) {
				l.Depth, _ = strconv.Atoi(f[i+1])
			}
		case "score":
			if i+2 < len(f) {
				n, err := strconv.Atoi(f[i+2])
				if err == nil && (f[i+1] == "cp" || f[i+1] == "mate") {
					l.Score = Score{f[i+1], n}
					score = true
				}
			}
		case "pv":
			l.Moves = append([]string{}, f[i+1:]...)
			return l, score && l.Depth > 0 && len(l.Moves) > 0
		}
	}
	return l, false
}
