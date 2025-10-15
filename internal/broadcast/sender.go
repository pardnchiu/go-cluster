package broadcast

import (
	"context"
	"encoding/json"
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
	peers    []node.Peer
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

	node.Local.Mu.RLock()
	defer node.Local.Mu.RUnlock()

	return &sender{
		conn:     conn,
		interval: time.Duration(sec) * time.Second,
		peers:    node.Local.Peers,
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
			if len(s.peers) == 0 {
				slog.Warn("no node to broadcast")
				continue
			}

			for _, e := range s.peers {
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
