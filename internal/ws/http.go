package ws

import (
	"encoding/json"
	"net/http"
)

func ListPlayersHandler(w http.ResponseWriter, r *http.Request) {
	players := ListPlayers()

	w.Header().Set("Content-Type", "application/json")

	w.Header().Set("Access-Control-Allow-Origin", "*")

	b, err := json.Marshal(players)
	if err != nil {
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func ListGamesHandler(w http.ResponseWriter, r *http.Request) {
	type gameSummary struct {
		MatchID   string   `json:"match_id"`
		PlayerAID string   `json:"playerA_id"`
		PlayerBID string   `json:"playerB_id"`
		ReadyA    bool     `json:"readyA"`
		ReadyB    bool     `json:"readyB"`
		Turn      string   `json:"turn"`
		Players   []string `json:"players"`
	}

	gamesMu.RLock()
	out := make([]gameSummary, 0, len(games))
	for id, g := range games {
		g.mu.Lock()
		readyA := g.Ready[g.PlayerAID]
		readyB := g.Ready[g.PlayerBID]
		turn := string(g.Turn)
		players := []string{g.PlayerAID, g.PlayerBID}
		g.mu.Unlock()

		out = append(out, gameSummary{
			MatchID:   id,
			PlayerAID: g.PlayerAID,
			PlayerBID: g.PlayerBID,
			ReadyA:    readyA,
			ReadyB:    readyB,
			Turn:      turn,
			Players:   players,
		})
	}
	gamesMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		http.Error(w, "internal_error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
