package raft

import (
	"math/rand"
	"time"
)

func genRand(min, max int) int {
	rad := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rad.Intn(max-min) + min
}
