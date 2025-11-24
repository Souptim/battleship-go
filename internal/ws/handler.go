package ws

import (
	"net/http"

	"github.com/google/uuid"
)

func HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	p := &Player{
		ID:   uuid.NewString(),
		Name: "",
		conn: conn,
		send: make(chan []byte, 256),
	}

	go p.writePump()
	go p.readPump()

	RegisterPlayer(p)

	welcome := `{"type":"welcome","id":"` + p.ID + `"}`
	p.send <- []byte(welcome)
}
