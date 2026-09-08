package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"tinychess/internal/game"
)

// The sender field doubles as an anonymous credential in incoming commands.
// Only its owner may receive it back. Other viewers receive a game-scoped alias
// that cannot authorize moves. Preserve the owner's echo for client deduping.
func eventForClient(data []byte, gameID, clientID string) []byte {
	var reaction game.ReactionPayload
	if json.Unmarshal(data, &reaction) != nil || reaction.Kind != "emoji" || reaction.Sender == clientID {
		return data
	}
	sum := sha256.Sum256([]byte(gameID + "\x00" + reaction.Sender))
	reaction.Sender = "public:" + hex.EncodeToString(sum[:])
	encoded, _ := json.Marshal(reaction)
	return encoded
}
