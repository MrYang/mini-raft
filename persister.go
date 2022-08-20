package raft

import "sync"

type Persister interface {
	SaveState(data []byte)
	ReadState() []byte
	SaveSnapshot(data []byte)
	ReadSnapshot() []byte
}

type MemoryPersister struct {
	sync.Mutex
	State    []byte
	Snapshot []byte
}

func NewMemoryPersister() *MemoryPersister {
	return &MemoryPersister{}
}

func (p *MemoryPersister) SaveState(data []byte) {
	p.Lock()
	defer p.Unlock()
	p.State = data
}

func (p *MemoryPersister) ReadState() []byte {
	p.Lock()
	defer p.Unlock()
	return p.State
}

func (p *MemoryPersister) SaveSnapshot(data []byte) {
	p.Lock()
	defer p.Unlock()
	p.Snapshot = data
}

func (p *MemoryPersister) ReadSnapshot() []byte {
	p.Lock()
	defer p.Unlock()
	return p.Snapshot
}
