package raft

import (
	_ "embed"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Servers []ServerConfig `toml:"servers"`
}

type ServerConfig struct {
	ID         int    `toml:"id"`
	Listen     string `toml:"listen"`
	HttpListen string `toml:"http_listen"`
}

var GlobalConfig Config

//go:embed config.toml
var strConfig string

func init() {
	_, err := toml.Decode(strConfig, &GlobalConfig)
	if err != nil {
		panic(err)
	}
}

func findConfig(ID int) (c ServerConfig) {
	for _, s := range GlobalConfig.Servers {
		if s.ID == ID {
			c = s
			break
		}
	}
	return
}
