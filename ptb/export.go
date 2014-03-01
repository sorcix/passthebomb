package ptb

import (
	"encoding/json"
)

const (
	jsonPrefix = ""
	jsonIndent = "\t"
)

// Export returns game details as JSON.
func Export(g *Game) ([]byte, error) {
	return json.MarshalIndent(g, jsonPrefix, jsonIndent)
}
