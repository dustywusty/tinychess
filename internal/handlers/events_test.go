package handlers

import (
	"bytes"
	"encoding/json"
	"testing"

	"tinychess/internal/game"
)

func TestReactionDoesNotExposeAnonymousCredential(t *testing.T) {
	secret := "private-player-credential"
	original, _ := json.Marshal(game.ReactionPayload{Kind: "emoji", Emoji: "👏", Sender: secret, At: 123})
	public := eventForClient(original, "game-a", "spectator")
	if bytes.Contains(public, []byte(secret)) {
		t.Fatal("reaction leaked the player credential")
	}
	if !bytes.Equal(original, eventForClient(original, "game-a", secret)) {
		t.Fatal("owner echo no longer deduplicates")
	}
	if !bytes.Equal(public, eventForClient(original, "game-a", "friend")) {
		t.Fatal("public sender is not stable within the game")
	}
	if bytes.Equal(public, eventForClient(original, "game-b", "friend")) {
		t.Fatal("alias correlates identities across games")
	}
	var reaction game.ReactionPayload
	if err := json.Unmarshal(public, &reaction); err != nil {
		t.Fatal(err)
	}
	if reaction.Emoji != "👏" || reaction.At != 123 {
		t.Fatal("reaction content changed")
	}
	state := []byte(`{"kind":"state","uci":["e2e4"]}`)
	if !bytes.Equal(state, eventForClient(state, "game-a", "viewer")) {
		t.Fatal("position event changed")
	}
}
