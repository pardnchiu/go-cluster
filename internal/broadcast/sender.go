package broadcast

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"time"

	"github.com/pardnchiu/pdcluster/internal/node"
	"github.com/pardnchiu/pdcluster/internal/util"
)

type sender struct {
	conn     *net.UDPConn
	interval time.Duration
	nodes    []string
}

func InitSender() (*sender, error) {
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

	return &sender{
		conn:     conn,
		interval: time.Duration(sec) * time.Second,
		// TODO: add sub nodes management, and cluster config to store peers
		nodes: []string{"10.7.22.252:7989"},
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
			data, err := node.CheckHealth()
			if err != nil {
				slog.Error("failed: check health",
					slog.String("error", err.Error()),
				)
				continue
			}

			jsonData := data.([]byte)
			nodes := s.nodes
			if len(nodes) == 0 {
				slog.Warn("no node to broadcast")
				continue
			}

			for _, e := range nodes {
				addr, err := net.ResolveUDPAddr("udp4", e)
				if err != nil {
					slog.Error("invalid: address",
						slog.String("node", e),
						slog.String("error", err.Error()),
					)
					continue
				}

				if _, err := s.conn.WriteToUDP(jsonData, addr); err != nil {
					slog.Error("failed: send data",
						slog.String("node", e),
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
