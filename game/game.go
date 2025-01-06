package game

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type Question struct {
	Clue     string `json:"clue"`
	Answer   string `json:"answer"`
	Points   int    `json:"points"`
	Category string `json:"category"`
	Col      int    `json:"col"`
	Row      int    `json:"row"`
	// Opened bool   `json:"opened"`
}

type GameState struct {
	Questions []Question `json:"questions"`
	Category  string     `json:"category"`
}

func InitGameState(r *http.Request) (*[]GameState, error) {
	var data []GameState

	defer r.Body.Close()

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		slog.Error("Couldn't read game state", "body", string(bytes[0:20]))
		return nil, err
	}

	err = json.Unmarshal(bytes, &data)

	if err != nil {
		slog.Error("Couldn't unmarshal game state", "body", string(bytes[0:20]), "err", err)
		return nil, err
	}

	return &data, nil
}
