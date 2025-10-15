package broadcast

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/pardnchiu/pdcluster/internal/node"
	"github.com/pardnchiu/pdcluster/internal/util"
)

type sender struct {
	conn     *net.UDPConn
	interval time.Duration
	node     *node.Node
}

func InitSender() (*sender, error) {
	if node.Local == nil {
		return nil, fmt.Errorf("failed: node not initialized")
	}
	sec := util.GetEnv("BROADCAST_INTERVAL").Int(5)
	// * udp can use any available port
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	node.Local.Mu.RLock()
	defer node.Local.Mu.RUnlock()

	return &sender{
		conn:     conn,
		interval: time.Duration(sec) * time.Second,
		node:     node.Local,
	}, nil
}

func (s *sender) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			data, err := s.node.CheckHealth()
			if err != nil {
				slog.Error("failed: check health",
					slog.String("error", err.Error()),
				)
				continue
			}

			jsonData := data.([]byte)
			if len(s.node.Peers) == 0 {
				slog.Warn("no node to broadcast")
				continue
			}

			for _, e := range s.node.Peers {
				node := e.IP + ":" + strconv.Itoa(e.Port)
				addr, err := net.ResolveUDPAddr("udp4", node)
				if err != nil {
					slog.Error("invalid: address",
						slog.String("node", node),
						slog.String("error", err.Error()),
					)
					continue
				}

				if _, err := s.conn.WriteToUDP(jsonData, addr); err != nil {
					slog.Error("failed: send data",
						slog.String("node", node),
						slog.String("error", err.Error()),
					)
					continue
				}
			}

			var health node.Health
			if err := json.Unmarshal(jsonData, &health); err != nil {
				slog.Error("failed: unmarshal json",
					slog.String("error", err.Error()),
				)
				continue
			}
		}
	}
}

func (s *sender) Close() error {
	return s.conn.Close()
}
