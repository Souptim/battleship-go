package main

import (
	"battleship-go/internal/ws"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.HandleWS)
	mux.HandleFunc("/api/players", ws.ListPlayersHandler)
	mux.HandleFunc("/api/games", ws.ListGamesHandler)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server Running on :%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
