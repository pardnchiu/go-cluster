package node

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pardnchiu/pdcluster/internal/util"
)

const (
	AppName = "go-cluster"
)

var (
	Local = &Node{
		Mu:    sync.RWMutex{},
		Node:  "",
		IP:    "",
		Peers: []Peer{},
	}
	configFolder   = fmt.Sprintf("/etc/%s", AppName)
	configFilepath = filepath.Join(configFolder, "conf.json")
)

type Node struct {
	Mu    sync.RWMutex
	Node  string
	IP    string
	Peers []Peer
}

type Config struct {
	Node  string `json:"node"`
	IP    string `json:"ip"`
	Peers []Peer `json:"peers"`
}

type Peer struct {
	Node    string `json:"node"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	Seq     int    `json:"seq"`
	Health  bool   `json:"health"`
	Removed bool   `json:"removed"`
}

func (n *Node) Init() error {
	if err := n.KeyPair(); err != nil {
		return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 0, err)
	}

	port := util.GetEnv("RECEIVE_PORT").Int(7989)

	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(configFolder, 0755); err != nil {
			return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 1, err)
		}
	}

	if _, err := os.Stat(configFilepath); !os.IsNotExist(err) {
		data, err := os.ReadFile(configFilepath)
		if err != nil {
			return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 2, err)
		}
		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 3, err)
		}

		Local.IP = config.IP
		Local.Node = config.Node
		Local.Peers = config.Peers

		return nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = AppName
	}

	ip := ip()
	peers := []Peer{
		{
			Node:    hostname,
			IP:      ip,
			Port:    port,
			Seq:     0,
			Health:  true,
			Removed: false,
		},
	}

	config := Config{
		Node:  hostname,
		IP:    ip,
		Peers: peers,
	}

	Local.IP = config.IP
	Local.Node = config.Node
	Local.Peers = config.Peers

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 4, err)
	}

	if err := os.WriteFile(configFilepath, data, 0644); err != nil {
		return fmt.Errorf("[%s-%d: %w]", util.GetFuncName(), 5, err)
	}

	return nil
}
