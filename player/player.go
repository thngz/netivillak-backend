package player

import (
	"log/slog"

	"github.com/gorilla/websocket"
)

type Player struct {
	Nickname string
	Points   int
	Conn     *websocket.Conn
}

func (p *Player) SendMessage(msg interface{}) {
    slog.Info("Sending message to player", "player", p.Nickname)
	p.Conn.WriteJSON(msg)
}
