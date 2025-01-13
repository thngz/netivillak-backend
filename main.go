package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"netivillak/lobby"
	"netivillak/message"
	"netivillak/player"
	"netivillak/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // ignore origin for dev
}

var lobbies = lobby.InitLobbies()

func createConnection(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("Couldn't add connection!", "err", err)
		return nil
	}
	slog.Info("Created connection", "origin", r.Header["Origin"])
	return conn
}

func createLobbyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	var lobbyRequestData lobby.InitLobbyRequest

	defer r.Body.Close()

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		slog.Error("Couldn't read game state", "body", string(bytes[0:20]))
	}

	err = json.Unmarshal(bytes, &lobbyRequestData)

	if err != nil {
		slog.Error("Couldn't unmarshal init lobby data", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Couldn't create lobby"))
		return
	}

	code := utils.CreateRandomId(6)
	lobby := lobby.InitLobby(&lobbyRequestData)

	slog.Info("Creating new lobby", "lobby id", code)

	lobbies.Lobbies[code] = lobby
	w.WriteHeader(http.StatusOK)

	payload := struct {
		Code   string         `json:"code"`
		Player *player.Player `json:"player"`
	}{
		Code:   code,
		Player: lobby.Creator,
	}
	json.NewEncoder(w).Encode(message.InitSuccessResponse(payload, message.CREATED_LOBBY))
}

type JoinLobbyRequest struct {
	Code     string `json:"code"`
	Nickname string `json:"nickname"`
}

func joinLobbyHandler(w http.ResponseWriter, r *http.Request) {
	conn := createConnection(w, r)
	var req JoinLobbyRequest

	for {

		err := conn.ReadJSON(&req)
		if err != nil {
			slog.Warn("Read message error", "error", err)
			conn.WriteJSON(message.InitFailureResponse(err.Error()))
		}

		lobby, ok := lobbies.Lobbies[req.Code]

		if !ok {
			slog.Debug("Connection invalid lobby id", "connection", conn.RemoteAddr().String(), "id", req.Code)
			conn.WriteJSON(message.InitFailureResponse(fmt.Sprintf("Lobby with id %s not found!", req.Code)))
			continue
		}

		if lobby.Creator.Nickname == req.Nickname {
			lobby.Creator.Conn = conn
		} else {
			p := player.Init(req.Nickname)
			p.Conn = conn
			lobby.AddPlayer(p)
		}

		conn.WriteJSON(message.InitSuccessResponse("Successfully joined lobby", message.JOINED_LOBBY))
		slog.Debug("Connection ", conn.RemoteAddr().String(), "Joined successfully!")

		nicknames := make([]string, 0, len(lobby.Players))

		for _, p := range lobby.Players {
			nicknames = append(nicknames, p.Nickname)
		}
		broadcast(*lobby, message.InitSuccessResponse(nicknames, message.PLAYERS_JOINED))
	}
}

func broadcast(l lobby.Lobby, msg interface{}) {
	for _, p := range l.Players {
		p.SendMessage(msg)
	}
}

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/createlobby", createLobbyHandler).Methods("POST")
	r.HandleFunc("/joinlobby", joinLobbyHandler)

	http.Handle("/", r)

	slog.Info("Starting server on port :3000")
	err := http.ListenAndServe(":3000", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
