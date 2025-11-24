package game

import "time"

type Cell int

const (
	Empty Cell = iota
	Ship
	Hit
	Miss
)

type Board [10][10]Cell

type PlayerSide int

const (
	SideA PlayerSide = iota
	SideB
)

type MatchState int

const (
	WaitingForPlayers MatchState = iota
	InProgress
	Finished
)

type Player struct {
	ID   string
	Name string
}

type Match struct {
	ID      string
	PlayerA *Player
	PlayerB *Player
	BoardA  Board
	BoardB  Board
	Turn    PlayerSide
	State   MatchState
	Created time.Time
}
