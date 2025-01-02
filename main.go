package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"netivillak/utils"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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

type Lobby struct {
	conns        map[*websocket.Conn]bool
	initialState *[]GameState
	name         string
}

type Lobbies struct {
	lobbies map[string]*Lobby
}

type JoinRequest struct {
	Code string `json:"code"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // ignore origin for dev
}

func InitLobby() *Lobby {
	return &Lobby{
		conns: make(map[*websocket.Conn]bool),
	}
}

func (l *Lobby) AddConnection(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("Couldn't add connection!", "err", err)
		return nil
	}
	l.conns[conn] = true
	slog.Info("Added connection", "origin", r.Header["Origin"], "lobby", l.name)
	w.WriteHeader(http.StatusOK)
	conn.WriteMessage(1, []byte("Connected to lobby"))
    
    return conn
}

func createLobbyHandler(l *Lobbies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		lobby := InitLobby()

		state, err := getState(r)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Couldn't create lobby"))
			return
		}
		id := utils.CreateRandomId(6)
		slog.Info("Creating new lobby", "lobby id", id)
		lobby.initialState = state
		lobby.name = id

		l.lobbies[id] = lobby
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id))
	}
}

func joinLobbyHandler(l *Lobbies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data JoinRequest
		w.Header().Set("Access-Control-Allow-Origin", "*")

		defer r.Body.Close()

		bytes, err := io.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.Error("Couldn't read body of join lobby request", "body", string(bytes[0:20]), "err", err)
			w.Write([]byte(err.Error()))
			return
		}

		err = json.Unmarshal(bytes, &data)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.Error("Couldn't unmarshal join lobby request", "body", string(bytes[0:20]), "err", err)
			w.Write([]byte(err.Error()))
			return
		}

		lobby, ok := l.lobbies[data.Code]
		slog.Info(data.Code)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("No such lobby"))
			return
		}
		lobby.AddConnection(w, r)
	}
}

func getState(r *http.Request) (*[]GameState, error) {
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

func main() {

	r := mux.NewRouter()

	lobbies := &Lobbies{
		lobbies: make(map[string]*Lobby),
	}

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
