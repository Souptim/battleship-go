package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Player struct {
	ID   string
	Name string
	conn *websocket.Conn
	send chan []byte
}

func (p *Player) readPump() {
	defer func() {

		UnregisterPlayer(p.ID)

		p.conn.Close()
		log.Println("readPump: exiting for", p.ID)
	}()
	log.Println("readPump: starting for", p.ID)

	log.Println("readPump: starting for", p.ID)
	p.conn.SetReadLimit(512)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPongHandler(func(string) error { p.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		mt, message, err := p.conn.ReadMessage()
		if err != nil {
			log.Println("readPump: starting for", p.ID)
			break
		}
		if mt != websocket.TextMessage {
			continue
		}
		var envelope struct {
			Type string `json:"type"`
			Name string `json:"name,omitempty"`
		}
		if err := json.Unmarshal(message, &envelope); err != nil {
			continue
		}
		switch envelope.Type {
		case "join":
			if envelope.Name != "" {
				p.Name = envelope.Name
			} else {
				p.Name = "Player-" + p.ID[:8]
			}
			ack := map[string]string{
				"type": "join_ack",
				"id":   p.ID,
				"name": p.Name,
			}
			ackBytes, _ := json.Marshal(ack)
			p.send <- ackBytes
		case "challenge":
			var payload struct {
				TargetID string `json:"target_id"`
			}
			if err := json.Unmarshal(message, &payload); err != nil || payload.TargetID == "" {
				continue
			}
			if target, ok := GetPlayer(payload.TargetID); ok {
				req := map[string]string{
					"type":      "challenge_request",
					"from_id":   p.ID,
					"from_name": p.Name,
				}
				b, _ := json.Marshal(req)
				target.send <- b
			} else {
				errMsg := map[string]string{
					"type":  "error",
					"error": "target_not_found",
				}
				b, _ := json.Marshal(errMsg)
				p.send <- b
			}

		case "challenge_response":
			var payload struct {
				TargetID string `json:"target_id"`
				Accept   bool   `json:"accept"`
			}
			if err := json.Unmarshal(message, &payload); err != nil || payload.TargetID == "" {
				continue
			}

			challenger, ok := GetPlayer(payload.TargetID)
			if !ok {

				resp := map[string]string{
					"type":  "error",
					"error": "challenger_not_connected",
				}
				b, _ := json.Marshal(resp)
				p.send <- b
				continue
			}

			forward := map[string]interface{}{
				"type":      "challenge_response_forward",
				"from_id":   p.ID,
				"from_name": p.Name,
				"accept":    payload.Accept,
				"target_id": payload.TargetID,
			}
			fb, _ := json.Marshal(forward)
			challenger.send <- fb

			if payload.Accept {

				m, assignment := createMatch(p.ID, challenger.ID)

				RegisterMatchState(m)

				chalSide := assignment[challenger.ID]
				chalMsg := map[string]interface{}{
					"type":          "match_start",
					"match_id":      m.ID,
					"your_side":     string(chalSide),
					"opponent_id":   p.ID,
					"opponent_name": p.Name,
				}
				cb, _ := json.Marshal(chalMsg)
				challenger.send <- cb

				accSide := assignment[p.ID]
				accMsg := map[string]interface{}{
					"type":          "match_start",
					"match_id":      m.ID,
					"your_side":     string(accSide),
					"opponent_id":   challenger.ID,
					"opponent_name": challenger.Name,
				}
				ab, _ := json.Marshal(accMsg)
				p.send <- ab

			}
		case "place_ships":

			var payload struct {
				MatchID string          `json:"match_id"`
				Ships   []ShipPlacement `json:"ships"`
			}
			if err := json.Unmarshal(message, &payload); err != nil {

				errMsg := map[string]string{"type": "error", "error": "bad_place_ships"}
				b, _ := json.Marshal(errMsg)
				p.send <- b
				continue
			}

			if err := SetPlayerShips(payload.MatchID, p.ID, payload.Ships); err != nil {
				errMsg := map[string]string{"type": "ships_error", "error": err.Error()}
				b, _ := json.Marshal(errMsg)
				p.send <- b
				continue
			}

			okMsg := map[string]string{"type": "ships_ok", "match_id": payload.MatchID}
			b, _ := json.Marshal(okMsg)
			p.send <- b

		case "shot_fired":

			var payload struct {
				MatchID string `json:"match_id"`
				X       int    `json:"x"`
				Y       int    `json:"y"`
			}
			if err := json.Unmarshal(message, &payload); err != nil {
				log.Println("conn: bad_shot_payload", err)
				errMsg := map[string]string{"type": "error", "error": "bad_shot_payload"}
				b, _ := json.Marshal(errMsg)
				p.send <- b
				continue
			}
			log.Println("conn: shot_fired received from", p.ID, "payload:", payload)

			_, err := ProcessShot(payload.MatchID, p.ID, payload.X, payload.Y)
			if err != nil {
				errMsg := map[string]string{"type": "shot_error", "error": err.Error()}
				b, _ := json.Marshal(errMsg)
				p.send <- b
				continue
			}

		default:

		}
	}
	log.Println("readPump: exiting for", p.ID)
}

func (p *Player) writePump() {
	log.Println("writePump: starting for", p.ID)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.conn.Close()
		log.Println("writePump: exiting for", p.ID)
	}()

	for {
		select {
		case msg, ok := <-p.send:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				log.Println("writePump: send channel closed for", p.ID)
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := p.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
