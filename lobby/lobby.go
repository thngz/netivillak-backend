package lobby

import (
	"netivillak/game"

	"github.com/gorilla/websocket"
)

type Lobby struct {
	Conns        map[*websocket.Conn]bool
	InitialState *[]game.GameState
	Name         string
}

type Lobbies struct {
	Lobbies map[string]*Lobby
}

func InitLobby() *Lobby {
	return &Lobby{
		Conns: make(map[*websocket.Conn]bool),
	}
}

func InitLobbies() *Lobbies {
	return &Lobbies{
		Lobbies: make(map[string]*Lobby),
	}

}
