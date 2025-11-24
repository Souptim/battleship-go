package ws

import (
	"errors"

	"battleship-go/internal/game"
)

type ShipPlacement struct {
	Type string `json:"type"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Dir  string `json:"dir"`
}

var ShipSizes = map[string]int{
	"carrier":    5,
	"battleship": 4,
	"cruiser":    3,
	"submarine":  3,
	"destroyer":  2,
}

func BuildBoardFromPlacements(ships []ShipPlacement) (game.Board, error) {
	var b game.Board
	seen := map[string]int{}
	for _, s := range ships {
		size, ok := ShipSizes[s.Type]
		if !ok {
			return b, errors.New("unknown_ship_type:" + s.Type)
		}
		seen[s.Type]++

		if s.Dir == "H" {
			if s.X < 0 || s.Y < 0 || s.X+size-1 > 9 || s.Y > 9 {
				return b, errors.New("out_of_bounds:" + s.Type)
			}
			for i := 0; i < size; i++ {
				if b[s.Y][s.X+i] != game.Empty {
					return b, errors.New("overlap")
				}
				b[s.Y][s.X+i] = game.Ship
			}
		} else if s.Dir == "V" {
			if s.X < 0 || s.Y < 0 || s.Y+size-1 > 9 || s.X > 9 {
				return b, errors.New("out_of_bounds:" + s.Type)
			}
			for i := 0; i < size; i++ {
				if b[s.Y+i][s.X] != game.Empty {
					return b, errors.New("overlap")
				}
				b[s.Y+i][s.X] = game.Ship
			}
		} else {
			return b, errors.New("invalid_direction")
		}
	}

	for name, _ := range ShipSizes {
		if seen[name] != 1 {
			return b, errors.New("invalid_fleet")
		}
	}

	return b, nil
}
