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
	PlayerBotID   string `json:"playerBotId,omitempty"`
	MoveDelayMs   *int   `json:"moveDelayMs,omitempty"`
}

type BotSettings struct {
	PlayerBotID string `json:"playerBotId,omitempty"`
	MoveDelayMs *int   `json:"moveDelayMs,omitempty"`
}

func (s BotSettings) Valid() bool {
	return (s.PlayerBotID == "" || ValidBotID(s.PlayerBotID)) && (s.MoveDelayMs == nil || (*s.MoveDelayMs >= 0 && *s.MoveDelayMs <= 5000))
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
	return g.ConfigureBotWithSettings(id, owner, color, BotSettings{})
}

func (g *Game) ConfigureBotWithSettings(id, owner, color string, settings BotSettings) error {
	g.Mu.Lock()
	defer g.Mu.Unlock()
	if !ValidBotID(id) || !settings.Valid() || owner == "" || (color != "w" && color != "b") || len(g.Clients) != 0 || len(g.g.Moves()) != 0 {
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
	delay := settings.MoveDelayMs
	if delay == nil && settings.PlayerBotID != "" {
		value := 1500
		delay = &value
	}
	g.Bot = &BotOpponent{ID: id, Color: botColor, PolicyVersion: BotPolicyVersion, PlayerBotID: settings.PlayerBotID, MoveDelayMs: delay}
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
		if g.Bot.PlayerBotID != "" {
			side = g.g.Position().Turn()
		}
		return g.makeMoveForColorLocked(side, request.UCI)
	}
	if g.Bot != nil && g.Bot.PlayerBotID != "" {
		return "", fmt.Errorf("both sides are controlled by computers")
	}
	side, ok := g.Clients[request.ClientID]
	if !ok {
		return "", fmt.Errorf("unknown client")
	}
	return g.makeMoveForColorLocked(side, request.UCI)
}
