package broadcast

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"

	"github.com/pardnchiu/pdcluster/internal/database"
	"github.com/pardnchiu/pdcluster/internal/node"
	"github.com/pardnchiu/pdcluster/internal/util"
)

type Receiver struct {
	conn *net.UDPConn
	db   *database.SQLite
}

func InitReceiver(db *database.SQLite) (*Receiver, error) {
	port := util.GetEnv("RECEIVE_PORT").Int(7989)
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: port,
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	return &Receiver{
		conn: conn,
		db:   db,
	}, nil
}

func (r *Receiver) Start(ctx context.Context) error {
	// {
	// 	"cpu": 0.0124,
	// 	"disk": 4517707776,
	// 	"id": "qemu",
	// 	"maxcpu": 4,
	// 	"maxdisk": 8242348032,
	// 	"maxmem": 4104863744,
	// 	"mem": 2371907584,
	// 	"ip": "10.7.22.252",
	// 	"status": "online",
	// 	"uptime": 35835
	// }
	// around 200~300 bytes
	// use 320 bytes for safety
	buf := make([]byte, 320)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		bytes, addr, err := r.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				slog.Error("failed: read from udp",
					slog.String("error", err.Error()),
				)
				continue
			}
		}

		var data node.Health
		// println("Received data", string(buf[:bytes]))
		if err := json.Unmarshal(buf[:bytes], &data); err != nil {
			slog.Error("failed: unmarshal json",
				slog.String("from", addr.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := r.db.UpdateHealth(&data); err != nil {
			slog.Error("failed: update health",
				slog.String("error", err.Error()),
			)
			continue
		}

		slog.Info("success: updated health",
			slog.String(addr.String(), data.Status),
		)
	}
}

func (r *Receiver) Close() error {
	return r.conn.Close()
}
