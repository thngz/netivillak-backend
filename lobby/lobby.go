package lobby

import (
	"netivillak/game"
	"netivillak/player"
)

type InitLobbyRequest struct {
	GameRows        []game.GameRow `json:"gameRows"`
	CreatorNickName string         `json:"creatorNickname"`
}

type Lobby struct {
	Players      []*player.Player
	Creator      *player.Player
	InitialState *[]game.GameRow
	Name         string
}

func (l *Lobby) AddPlayer(player *player.Player) {
	l.Players = append(l.Players, player)
}

func InitLobby(r *InitLobbyRequest) *Lobby {
	c := player.Init(r.CreatorNickName)
    
	l := &Lobby{
		Creator: c,
	}

	l.AddPlayer(c)
	return l
}

type Lobbies struct {
	Lobbies map[string]*Lobby
}

func InitLobbies() *Lobbies {
	return &Lobbies{
		Lobbies: make(map[string]*Lobby),
	}

}
