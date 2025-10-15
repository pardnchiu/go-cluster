package node

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pardnchiu/pdcluster/internal/util"
)

var (
	Local   Cluster
	AppName = "go-cluster"
)

var (
	configFolder = "/etc/go-cluster"
	configName   = "conf.json"
	configPath   = filepath.Join(configFolder, configName)
)

type Cluster struct {
	Mu    sync.RWMutex
	Node  string `json:"node"`
	IP    string `json:"ip"`
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

func InitConfig() error {
	port := util.GetEnv("RECEIVE_PORT").Int(7989)

	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(configFolder, 0755); err != nil {
			return fmt.Errorf("[InitConfig-0: %w]", err)
		}
	}

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("[InitConfig-1: %w]", err)
		}
		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("[InitConfig-2: %w]", err)
		}

		Local = Cluster{
			Mu:    sync.RWMutex{},
			Node:  config.Node,
			IP:    config.IP,
			Peers: config.Peers,
		}

		return nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = AppName
	}

	ip := GetIP()
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

	Local = Cluster{
		Mu:    sync.RWMutex{},
		Node:  hostname,
		IP:    ip,
		Peers: peers,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("[InitConfig-2: %w]", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("[InitConfig-3: %w]", err)
	}

	return nil
}
