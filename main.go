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

	"netivillak/game"
	"netivillak/lobby"
	"netivillak/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // ignore origin for dev
}

func createConnection(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("Couldn't add connection!", "err", err)
		return nil
	}
	slog.Info("Created connection", "origin", r.Header["Origin"])
	conn.WriteMessage(1, []byte("Connection made"))

	return conn
}

func createLobbyHandler(l *lobby.Lobbies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		lobby := lobby.InitLobby()

		state, err := getState(r)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Couldn't create lobby"))
			return
		}
		id := utils.CreateRandomId(6)
		slog.Info("Creating new lobby", "lobby id", id)
		lobby.InitialState = state
		lobby.Name = id

		l.Lobbies[id] = lobby
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id))
	}
}

func joinLobbyHandler(l *lobby.Lobbies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn := createConnection(w, r)

		for {
			_, bytes, err := conn.ReadMessage()

			if err != nil {
				slog.Warn("Read message error", "error", err)
			}

			id := string(bytes)
			lobby, ok := l.Lobbies[id]

			if !ok {
				conn.WriteMessage(1, []byte("Invalid id!"))
				// conn.Close()
				slog.Warn("Connection invalid lobby id", "connection", conn.RemoteAddr().String(), "id", id)
				continue
			}

			lobby.Conns[conn] = true
			slog.Info("Connection ", conn.RemoteAddr().String(), "Joined successfully!")
		}
	}
}

func getState(r *http.Request) (*[]game.GameState, error) {
	var data []game.GameState

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

func main() {

	r := mux.NewRouter()

	lobbies := lobby.InitLobbies()
	r.HandleFunc("/createlobby", createLobbyHandler(lobbies)).Methods("POST")
	r.HandleFunc("/joinlobby", joinLobbyHandler(lobbies))

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
