package raft

import (
	"bytes"
	"context"
	"encoding/gob"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	Follower  = "follower"
	Candidate = "candidate"
	Leader    = "leader"
)

const (
	HeartBeatTime      = 100 * time.Millisecond
	electionTimeoutMin = 200
	electionTimeoutMax = 400
)

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

	// 持久化及传输接口
	persister   Persister
	transporter Transporter

	// 需要持久化
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

type RequestVoteArgs struct {
	Term         int
	CandidateID  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Entry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term          int
	Success       bool
	ConflictTerm  int
	ConflictIndex int
}

func (r *Raft) RequestVote(ctx context.Context, args *RequestVoteArgs, reply *RequestVoteReply) error {
	r.Lock()
	defer r.Unlock()

	log.Debugf("当前server %d, 当前任期:%d,  候选者 %d 请求投票, 参数:%+v", r.me, r.currentTerm, args.CandidateID, args)

	var lastLogIndex, lastLogTerm int

	rejectFunc := func() {
		reply.Term = r.currentTerm
		reply.VoteGranted = false
	}

	rejectByLogIndexFunc := func() bool {
		lastLogIndex = len(r.log)
		lastLogTerm = 0
		if lastLogIndex > 0 {
			lastLogTerm = r.log[lastLogIndex-1].Term
		}
		if args.LastLogTerm < lastLogTerm {
			rejectFunc()
			return true
		}
		if args.LastLogTerm == lastLogTerm {
			if args.LastLogIndex < lastLogIndex {
				rejectFunc()
				return true
			}
		}
		return false
	}

	agreeFunc := func() {
		reply.Term = r.currentTerm
		reply.VoteGranted = true

		r.votedFor = args.CandidateID
		r.Persist()
		r.setGrantVoteCh()
	}

	if args.Term < r.currentTerm {
		rejectFunc()
	} else if args.Term == r.currentTerm {
		if r.votedFor != -1 && r.votedFor != args.CandidateID {
			rejectFunc()
			goto end
		}
		if rejectByLogIndexFunc() {
			goto end
		}

		log.Debugf("当前server:%d 投票到候选者:%d", r.me, args.CandidateID)
		agreeFunc()
	} else {
		r.convertToFollower(args.Term, -1)
		if rejectByLogIndexFunc() {
			goto end
		}

		log.Debugf("当前server:%d 投票到候选者:%d", r.me, args.CandidateID)
		agreeFunc()
	}
end:
	log.Debugf("候选者 %d 请求投票结束, 参数:%+v, 返回值:%+v", args.CandidateID, args, reply)
	return nil
}

func (r *Raft) GetState(term int, isLeader bool) {
	r.Lock()
	defer r.Unlock()

	term = r.currentTerm
	if r.state == Leader {
		isLeader = true
	}
	return
}

func (r *Raft) Persist() {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	e.Encode(r.currentTerm)
	e.Encode(r.votedFor)
	e.Encode(r.log)
	r.persister.SaveState(b.Bytes())
}

func (r *Raft) readPersist(data []byte) {
	if len(data) == 0 {
		return
	}

	b := new(bytes.Buffer)
	d := gob.NewDecoder(b)
	d.Decode(&r.currentTerm)
	d.Decode(&r.votedFor)
	d.Decode(&r.log)
}

func (r *Raft) convertToFollower(term, voteFor int) {
	r.currentTerm = term
	r.state = Follower
	r.totalVotes = 0
	r.votedFor = voteFor
	r.Persist()
}

func (r *Raft) convertToCandidate() {
	r.state = Candidate
	r.currentTerm++
	r.votedFor = r.me
	r.totalVotes = 1
	r.electionTimeout = genRand(electionTimeoutMin, electionTimeoutMax)
	r.timer.Reset(time.Duration(r.electionTimeout) * time.Millisecond)
	r.Persist()
}

func (r *Raft) convertToLeader() {
	r.state = Leader
	r.nextIndex = make([]int, len(GlobalConfig.Servers))
	r.matchIndex = make([]int, len(GlobalConfig.Servers))
	for i := 0; i < len(GlobalConfig.Servers); i++ {
		r.nextIndex[i] = len(r.log) + 1
		r.matchIndex[i] = 0
	}
}

func (r *Raft) setHeartBeatCh() {
	go func() {
		select {
		case <-r.heartBeatch:
		default:
		}
		r.heartBeatch <- true
	}()
}

func (r *Raft) setGrantVoteCh() {
	go func() {
		select {
		case <-r.grantVoteCh:
		default:
		}
		r.grantVoteCh <- true
	}()
}

func (r *Raft) setLeaderCh() {
	go func() {
		select {
		case <-r.leaderCh:
		default:
		}
		r.leaderCh <- true
	}()
}
