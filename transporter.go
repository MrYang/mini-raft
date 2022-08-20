package raft

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
)

type Transporter interface {
	SendRequestVote(serverID int, req *RequestVoteArgs, reply *RequestVoteReply) error
	SendAppendEntries(serverID int, req *AppendEntriesArgs, reply *AppendEntriesReply) error
}

type RpcxTransporter struct {
	rpcClients map[int]client.XClient
}

func (r RpcxTransporter) SendRequestVote(serverID int, req *RequestVoteArgs, reply *RequestVoteReply) error {
	if xClient, ok := r.rpcClients[serverID]; ok {
		return xClient.Call(context.Background(), "RequestVote", req, reply)
	}
	return errors.New("调用投票rpc 失败")
}

func (r RpcxTransporter) SendAppendEntries(serverID int, req *AppendEntriesArgs, reply *AppendEntriesReply) error {
	if xClient, ok := r.rpcClients[serverID]; ok {
		return xClient.Call(context.Background(), "AppendEntries", req, reply)
	}
	return errors.New("调用日志复制rpc 失败")
}

func NewRpcxTransporter(raft *Raft) *RpcxTransporter {
	transporter := &RpcxTransporter{rpcClients: make(map[int]client.XClient)}
	transporter.startRpcx(raft)
	return transporter
}

func (r RpcxTransporter) startRpcx(raft *Raft) {
	go r.initRpcxServer(raft)
	go r.initRpcxClient(raft)
}

func (r RpcxTransporter) initRpcxServer(raft *Raft) {
	s := server.NewServer()
	err := s.Register(raft, "")
	if err != nil {
		panic(err)
	}

	config := findConfig(raft.me)
	log.Info("start rpcx listen:%s", config.Listen)

	err = s.Serve("tcp", config.Listen)
	if err != nil {
		panic(err)
	}
}

func (r RpcxTransporter) initRpcxClient(raft *Raft) {
	for _, sc := range GlobalConfig.Servers {
		if raft.me == sc.ID {
			continue
		}

		d, err := client.NewPeer2PeerDiscovery("tcp@127.0.0.1"+sc.Listen, "")
		if err != nil {
			log.Warnf("连接rpc 服务失败:%s", "tcp@127.0.0.1"+sc.Listen)
		}
		r.rpcClients[sc.ID] = client.NewXClient("Raft", client.Failtry, client.RandomSelect, d, client.DefaultOption)
	}
}
