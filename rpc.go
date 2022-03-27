package raft

import (
	"github.com/smallnest/rpcx/client"
)

var rpcClients map[int]client.XClient
