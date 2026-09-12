package game

import (
	"fmt"
	"github.com/corentings/chess/v2"
)

const BotPolicyVersion = 1

type BotOpponent struct {
	ID            string `json:"id"`
	Color         string `json:"color"`
	PolicyVersion int    `json:"policyVersion"`
}

func ValidBotID(id string) bool {
	switch id {
	case "pip", "max", "ada", "viktor", "machine":
		return true
	}
	return false
}

// ConfigureBot reserves the other side without creating a public bot credential.
func (g *Game) ConfigureBot(id, owner, color string) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if !ValidBotID(id) || owner == "" || (color != "w" && color != "b") || len(g.Clients) != 0 || len(g.g.Moves()) != 0 {
		return fmt.Errorf("invalid computer game")
	}
	g.OwnerID = owner
	g.OwnerColor = chess.White
	botColor := "b"
	if color == "b" {
		g.OwnerColor = chess.Black
		botColor = "w"
	}
	g.Clients[owner] = g.OwnerColor
	g.Bot = &BotOpponent{ID: id, Color: botColor, PolicyVersion: BotPolicyVersion}
	return nil
}

// SubmitMove keeps authorization and stale-position checks inside the same lock
// as legality and history. Multiple tabs can compute, but only one can commit.
func (g *Game) SubmitMove(request MoveRequest) (string, error) {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if request.ExpectedPly != nil && *request.ExpectedPly != len(g.g.Moves()) {
		return "", fmt.Errorf("position changed")
	}
	if request.BotMove {
		if g.Bot == nil || request.ClientID != g.OwnerID || request.ExpectedPly == nil {
			return "", fmt.Errorf("computer move not permitted")
		}
		side := chess.White
		if g.Bot.Color == "b" {
			side = chess.Black
		}
		return g.makeMoveForColorLocked(side, request.UCI)
	}
	side, ok := g.Clients[request.ClientID]
	if !ok {
		return "", fmt.Errorf("unknown client")
	}
	return g.makeMoveForColorLocked(side, request.UCI)
}
