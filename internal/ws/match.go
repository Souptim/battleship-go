package ws

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type Side string

const (
	SideA Side = "A"
	SideB Side = "B"
)

type Match struct {
	ID         string
	PlayerAID  string
	PlayerBID  string
	CreatedAt  time.Time
	StartedAt  time.Time
	AssignedAt time.Time
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// createMatch creates a match with random side assignment and returns the match plus a mapping
// telling for each player id which side they were assigned.
func createMatch(aID, bID string) (*Match, map[string]Side) {
	m := &Match{
		ID:        uuid.NewString(),
		PlayerAID: aID,
		PlayerBID: bID,
		CreatedAt: time.Now(),
	}

	// random assignment
	var mapping map[string]Side = make(map[string]Side)
	if rng.Intn(2) == 0 {

		// a is A, b is B
		mapping[aID] = SideA
		mapping[bID] = SideB
	} else {
		// swap
		mapping[aID] = SideB
		mapping[bID] = SideA
	}

	m.AssignedAt = time.Now()
	return m, mapping
}
