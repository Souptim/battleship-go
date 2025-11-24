package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"battleship-go/internal/game"
)

type GameState struct {
	MatchID    string
	PlayerAID  string
	PlayerBID  string
	Boards     map[string]game.Board
	Ready      map[string]bool
	Turn       Side
	mu         sync.Mutex
	ShipCells  map[string]map[string]map[string]bool
	ShipHealth map[string]map[string]int
}

var (
	gamesMu sync.RWMutex
	games   = make(map[string]*GameState)
)

func RegisterMatchState(m *Match) {
	g := &GameState{
		MatchID:   m.ID,
		PlayerAID: m.PlayerAID,
		PlayerBID: m.PlayerBID,
		Boards:    map[string]game.Board{},
		Ready:     map[string]bool{m.PlayerAID: false, m.PlayerBID: false},
	}

	g.ShipCells = make(map[string]map[string]map[string]bool)
	g.ShipHealth = make(map[string]map[string]int)

	g.ShipCells[m.PlayerAID] = make(map[string]map[string]bool)
	g.ShipCells[m.PlayerBID] = make(map[string]map[string]bool)

	g.ShipHealth[m.PlayerAID] = make(map[string]int)
	g.ShipHealth[m.PlayerBID] = make(map[string]int)

	gamesMu.Lock()
	games[m.ID] = g
	gamesMu.Unlock()
	log.Println("RegisterMatchState: created game state for match", m.ID, "players:", m.PlayerAID, m.PlayerBID)
}

func GetGameState(matchID string) (*GameState, bool) {
	gamesMu.RLock()
	defer gamesMu.RUnlock()
	g, ok := games[matchID]
	return g, ok
}

func SetPlayerShips(matchID, playerID string, placements []ShipPlacement) error {
	log.Println("SetPlayerShips called - match:", matchID, "player:", playerID, "placements:", len(placements))
	g, ok := GetGameState(matchID)
	if !ok {
		log.Println("SetPlayerShips: match not found:", matchID)
		return errors.New("match_not_found")
	}
	log.Println("SetPlayerShips: match players registered:", "playerA:", g.PlayerAID, "playerB:", g.PlayerBID)

	board, err := BuildBoardFromPlacements(placements)
	if err != nil {
		log.Println("SetPlayerShips: validation failed for player", playerID, "err:", err)
		return err
	}

	if g.ShipCells[playerID] == nil {
		g.ShipCells[playerID] = make(map[string]map[string]bool)
	}
	if g.ShipHealth[playerID] == nil {
		g.ShipHealth[playerID] = make(map[string]int)
	}

	for _, p := range placements {
		shipType := p.Type
		size, ok := ShipSizes[shipType]
		if !ok {
			log.Println("SetPlayerShips: unknown ship type (shouldn't happen):", shipType)
			continue
		}

		g.mu.Lock()
		if g.ShipCells[playerID][shipType] == nil {
			g.ShipCells[playerID][shipType] = make(map[string]bool)
		}
		g.mu.Unlock()

		for i := 0; i < size; i++ {
			rx := p.X
			ry := p.Y
			if p.Dir == "H" || p.Dir == "h" {
				rx = p.X + i
				ry = p.Y
			} else {
				rx = p.X
				ry = p.Y + i
			}
			if rx < 0 || ry < 0 || rx > 9 || ry > 9 {
				continue
			}
			key := fmt.Sprintf("%d_%d", rx, ry)

			g.mu.Lock()
			g.ShipCells[playerID][shipType][key] = true
			g.mu.Unlock()
		}

		g.mu.Lock()
		g.ShipHealth[playerID][shipType] = size
		g.mu.Unlock()

		log.Println("SetPlayerShips: populated ship", shipType, "for player", playerID, "size:", size)
	}

	g.mu.Lock()
	log.Println("DEBUG ShipHealth for player", playerID, ":", g.ShipHealth[playerID])
	g.mu.Unlock()
	g.mu.Lock()
	g.Boards[playerID] = board
	g.Ready[playerID] = true
	readyA := g.Ready[g.PlayerAID]
	readyB := g.Ready[g.PlayerBID]
	g.mu.Unlock()
	log.Println("SetPlayerShips: stored board for player", playerID, "readyA:", readyA, "readyB:", readyB)

	if readyA && readyB {
		log.Println("SetPlayerShips: both players ready for match", matchID)
		if rng != nil {
			if rng.Intn(2) == 0 {
				g.Turn = SideA
			} else {
				g.Turn = SideB
			}
		} else {
			g.Turn = SideA
		}

		notify := func(pID, opponentID string, yourSide Side) {
			msg := map[string]interface{}{
				"type":        "all_ships_ready",
				"match_id":    matchID,
				"start_turn":  string(g.Turn),
				"your_side":   string(yourSide),
				"opponent_id": opponentID,
			}
			b, _ := json.Marshal(msg)
			if pl, ok := GetPlayer(pID); ok {
				pl.send <- b
			}
		}

		notify(g.PlayerAID, g.PlayerBID, assignSideForPlayer(g, g.PlayerAID))
		notify(g.PlayerBID, g.PlayerAID, assignSideForPlayer(g, g.PlayerBID))
	}

	return nil
}

func assignSideForPlayer(g *GameState, playerID string) Side {
	if playerID == g.PlayerAID {
		return SideA
	}
	return SideB
}

func ProcessShot(matchID, shooterID string, x, y int) (map[string]interface{}, error) {
	log.Println("ProcessShot: ENTER match", matchID, "shooter", shooterID, "x", x, "y", y)
	g, ok := GetGameState(matchID)
	if !ok {
		log.Println("ProcessShot: match_not_found")
		return nil, errors.New("match_not_found")
	}

	if x < 0 || x > 9 || y < 0 || y > 9 {
		log.Println("ProcessShot: out_of_bounds")
		return nil, errors.New("out_of_bounds")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.Boards[g.PlayerAID]; !ok {
		return nil, errors.New("playerA_board_missing")
	}
	if _, ok := g.Boards[g.PlayerBID]; !ok {
		return nil, errors.New("playerB_board_missing")
	}

	var shooterSide Side
	if shooterID == g.PlayerAID {
		shooterSide = SideA
	} else if shooterID == g.PlayerBID {
		shooterSide = SideB
	} else {
		return nil, errors.New("unknown_player")
	}

	if g.Turn != shooterSide {
		log.Println("ProcessShot: not_your_turn. Turn:", g.Turn, "Shooter:", shooterSide)
		return nil, errors.New("not_your_turn")
	}

	var oppID string
	if shooterID == g.PlayerAID {
		oppID = g.PlayerBID
	} else {
		oppID = g.PlayerAID
	}

	board := g.Boards[oppID]

	cell := board[y][x]
	hit := false
	var sunkShip string

	result := map[string]interface{}{
		"type":       "shot_result",
		"match_id":   matchID,
		"x":          x,
		"y":          y,
		"shooter_id": shooterID,
		"hit":        hit,
	}

	log.Println("ProcessShot: match", matchID, "shooter", shooterID, "x", x, "y", y, "cell", cell)

	switch cell {
	case game.Ship:
		board[y][x] = game.Hit
		result["hit"] = true
		result["message"] = "hit"
		hit = true
		log.Println("ProcessShot: HIT!")

		key := fmt.Sprintf("%d_%d", x, y)
		if playerShips, ok := g.ShipCells[oppID]; ok {
			for shipType, cells := range playerShips {
				if cells == nil {
					continue
				}
				if cells[key] {

					remaining, has := g.ShipHealth[oppID][shipType]
					if !has {
						remaining = 0
						for _ = range cells {
							remaining++
						}
					}
					if remaining > 0 {
						g.ShipHealth[oppID][shipType] = remaining - 1
						log.Println("Ship hit:", shipType, "owner:", oppID, "remaining:", g.ShipHealth[oppID][shipType])
						if g.ShipHealth[oppID][shipType] == 0 {

							g.ShipHealth[oppID][shipType] = -1
							sunkShip = shipType
						}
					}
					break
				}
			}
		}

	case game.Empty:
		board[y][x] = game.Miss
		result["hit"] = false
		result["message"] = "miss"
		log.Println("ProcessShot: MISS!")
	case game.Hit, game.Miss:
		return nil, errors.New("already_shot")
	default:
		board[y][x] = game.Miss
		result["hit"] = false
		result["message"] = "miss"
		log.Println("ProcessShot: MISS (default)!")
	}

	g.Boards[oppID] = board

	oppShipsRemain := false
	for ry := 0; ry < 10; ry++ {
		for rx := 0; rx < 10; rx++ {
			if g.Boards[oppID][ry][rx] == game.Ship {
				oppShipsRemain = true
				break
			}
		}
		if oppShipsRemain {
			break
		}
	}

	if !oppShipsRemain {
		result["game_over"] = true
		result["winner_id"] = shooterID
	} else {

		if !hit {
			log.Println("ProcessShot: Switching turn. Current:", g.Turn)
			if g.Turn == SideA {
				g.Turn = SideB
			} else {
				g.Turn = SideA
			}
			log.Println("ProcessShot: New turn:", g.Turn)
		} else {
			log.Println("ProcessShot: Hit! Keeping turn:", g.Turn)
		}
		result["game_over"] = false
		result["next_turn"] = string(g.Turn)
	}

	result["target_id"] = oppID

	b, _ := json.Marshal(result)
	if pl, ok := GetPlayer(shooterID); ok {
		select {
		case pl.send <- b:
		default:
			log.Println("ProcessShot: shooter send blocked", shooterID)
		}
	}
	if pl, ok := GetPlayer(oppID); ok {
		select {
		case pl.send <- b:
		default:
			log.Println("ProcessShot: opponent send blocked", oppID)
		}
	}

	if sunkShip != "" {
		payload := map[string]interface{}{
			"type":      "ship_sunk",
			"match_id":  matchID,
			"ship_type": sunkShip,
			"owner_id":  oppID,
			"by_id":     shooterID,
		}
		pb, _ := json.Marshal(payload)
		if pl, ok := GetPlayer(shooterID); ok {
			select {
			case pl.send <- pb:
			default:
				log.Println("ProcessShot: shooter ship_sunk send blocked", shooterID)
			}
		}
		if pl, ok := GetPlayer(oppID); ok {
			select {
			case pl.send <- pb:
			default:
				log.Println("ProcessShot: opponent ship_sunk send blocked", oppID)
			}
		}
		log.Println("ship_sunk emitted:", sunkShip, "for match", matchID, "owner", oppID)
	}

	return result, nil
}
