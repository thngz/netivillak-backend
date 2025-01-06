package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"netivillak/game"
	"netivillak/lobby"
	"netivillak/message"
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

	lobby := lobby.InitLobby()

	state, err := game.InitGameState(r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Couldn't create lobby"))
		return
	}

	code := utils.CreateRandomId(6)
	slog.Info("Creating new lobby", "lobby id", code)
	lobby.InitialState = state
	lobby.Name = code

	lobbies.Lobbies[code] = lobby
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(message.InitSuccessResponse(code, message.CREATED_LOBBY))
}

func joinLobbyHandler(w http.ResponseWriter, r *http.Request) {
	conn := createConnection(w, r)

	for {
		_, bytes, err := conn.ReadMessage()

		if err != nil {
			slog.Warn("Read message error", "error", err)
			conn.WriteJSON(message.InitFailureResponse(err.Error()))
		}

		id := string(bytes)
		lobby, ok := lobbies.Lobbies[id]

		if !ok {
			slog.Warn("Connection invalid lobby id", "connection", conn.RemoteAddr().String(), "id", id)
			conn.WriteJSON(message.InitFailureResponse(fmt.Sprintf("Lobby with id %s not found!", id)))
			continue
		}

		lobby.Conns[conn] = true
		conn.WriteJSON(message.InitSuccessResponse("Successfully joined lobby", message.JOINED_LOBBY))
		slog.Info("Connection ", conn.RemoteAddr().String(), "Joined successfully!")

		keys := make([]string, 0, len(lobby.Conns))
		for c := range lobby.Conns {
			keys = append(keys, c.RemoteAddr().String())
		}

		broadcast(*lobby, message.InitSuccessResponse(keys, message.PLAYERS_JOINED))
	}
}

func broadcast(l lobby.Lobby, msg *message.SuccessResponse) {
	for c := range l.Conns {
		c.WriteJSON(msg)
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
