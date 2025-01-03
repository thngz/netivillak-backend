package game

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
