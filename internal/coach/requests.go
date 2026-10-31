package coach

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func requestFingerprint(r Request) string {
	data, _ := json.Marshal(r)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func replayResult(l *Lesson) Projection {
	p := project(l)
	p.Choices = l.LastChoices
	// Replaying an already committed action never starts another narration call.
	p.Explain = false
	return p
}
