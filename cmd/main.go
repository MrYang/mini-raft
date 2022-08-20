package main

import (
	_ "github.com/MrYang/mini-raft"
	raft "github.com/MrYang/mini-raft"
)

func main() {
	println(raft.GlobalConfig.Servers[0].Listen)
}
