package ws

import "sync"

var (
	playersMu sync.RWMutex
	players   = make(map[string]*Player)
)

// RegisterPlayer adds a player to the in-memory map.
func RegisterPlayer(p *Player) {
	playersMu.Lock()
	players[p.ID] = p
	playersMu.Unlock()
}

// UnregisterPlayer removes the player and closes the send channel.
func UnregisterPlayer(id string) {
	playersMu.Lock()
	if p, ok := players[id]; ok {
		// close send to signal writePump to exit
		close(p.send)
		delete(players, id)
	}
	playersMu.Unlock()
}

// GetPlayer returns the player by id (nil,false) if not found.
func GetPlayer(id string) (*Player, bool) {
	playersMu.RLock()
	p, ok := players[id]
	playersMu.RUnlock()
	return p, ok
}

// ListPlayers makes a lightweight snapshot of connected players (id + name).
func ListPlayers() []map[string]string {
	playersMu.RLock()
	defer playersMu.RUnlock()
	out := make([]map[string]string, 0, len(players))
	for _, p := range players {
		out = append(out, map[string]string{
			"id":   p.ID,
			"name": p.Name,
		})
	}
	return out
}
