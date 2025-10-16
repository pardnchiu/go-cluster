package database

import (
	"database/sql"
	"time"

	"github.com/pardnchiu/pdcluster/internal/node"
	_ "modernc.org/sqlite"
)

var (
	Local *SQLite
)

type SQLite sql.DB

// * templaure no other db
// type DB struct {
// 	SQLite *sql.DB
// }

func (d *SQLite) Init() error {
	// TODO: consider is put in /var/lib/ better than env to specify path?
	db, err := sql.Open("sqlite", "./health.db")
	if err != nil {
		return err
	}

	// * remove node, keep id
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS node_health (
	sn INTEGER PRIMARY KEY AUTOINCREMENT,
	cpu REAL,
	disk INTEGER,
	id TEXT NOT NULL UNIQUE,
	maxcpu INTEGER,
	maxdisk INTEGER,
	maxmem INTEGER,
	mem INTEGER,
	ip TEXT NOT NULL UNIQUE,
	status TEXT,
	uptime TEXT,
	last_update INTEGER,
	dismiss INTEGER DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_id ON node_health(id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ip ON node_health(ip);
CREATE INDEX IF NOT EXISTS idx_last_update ON node_health(last_update);
CREATE INDEX IF NOT EXISTS idx_status ON node_health(status);
`); err != nil {
		db.Close()
		return err
	}

	Local = (*SQLite)(db)

	return nil
}

func (db *SQLite) UpdateHealth(health *node.Health) error {
	_, err := (*sql.DB)(db).Exec(`
INSERT INTO node_health (
	cpu, disk, id, 
	maxcpu, maxdisk, maxmem, 
	mem, ip, status, 
	uptime, last_update, dismiss
) VALUES (
	?, ?, ?, 
	?, ?, ?, 
	?, ?, ?, 
	?, ?, 0
)
ON CONFLICT(id) DO UPDATE SET
	cpu = excluded.cpu,
	disk = excluded.disk,
	id = excluded.id,
	maxcpu = excluded.maxcpu,
	maxdisk = excluded.maxdisk,
	maxmem = excluded.maxmem,
	mem = excluded.mem,
	ip = excluded.ip,
	status = excluded.status,
	uptime = excluded.uptime,
	last_update = excluded.last_update
`,
		health.CPU, health.Disk, health.ID,
		health.MaxCPU, health.MaxDisk, health.MaxMem,
		health.Mem, health.IP, health.Status,
		health.Uptime, time.Now().Unix(),
	)

	return err
}

func (db *SQLite) Close() error {
	return (*sql.DB)(db).Close()
}
