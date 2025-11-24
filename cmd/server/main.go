package main

import (
	"battleship-go/internal/ws"
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.HandleWS)
	mux.HandleFunc("/api/players", ws.ListPlayersHandler)
	mux.HandleFunc("/api/games", ws.ListGamesHandler)
	mux.Handle("/", http.FileServer(http.Dir("web")))
	fmt.Println("Server Running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
