package raft

import (
	"sync"
	"time"
)

const (
	Follower  = "follower"
	Candidate = "candidate"
	Leader    = "leader"
)

const HeartBeatTime = 100 * time.Millisecond

type Entry struct {
	Term    int
	Command interface{}
}

type ApplyMsg struct {
	Index       int
	Command     interface{}
	UseSnapshot bool
	Snapshot    []byte
}

type Raft struct {
	sync.Mutex

	currentTerm int
	votedFor    int
	log         []Entry

	commitIndex int
	lastApplied int

	nextIndex  []int
	matchIndex []int

	me       int
	leaderID int

	state           string
	totalVotes      int
	electionTimeout int
	timer           *time.Timer
	applyCh         chan ApplyMsg
	grantVoteCh     chan bool
	heartBeatch     chan bool
	leaderCh        chan bool
}
