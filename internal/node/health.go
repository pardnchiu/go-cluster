package node

import (
	"encoding/json"
	"log/slog"
	"os"
	"runtime"
)

type Health struct {
	CPU     float64 `json:"cpu"`
	Disk    uint64  `json:"disk"`
	ID      string  `json:"id"`
	MaxCPU  int     `json:"maxcpu"`
	MaxDisk uint64  `json:"maxdisk"`
	MaxMem  uint64  `json:"maxmem"`
	Mem     uint64  `json:"mem"`
	IP      string  `json:"ip"`
	Status  string  `json:"status"`
	Uptime  int64   `json:"uptime"`
}

func CheckHealth() (interface{}, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	cpu, err := getCpuUsage()
	if err != nil {
		return nil, err
	}

	diskUsed, diskMax, err := getDisk("/")
	if err != nil {
		return nil, err
	}

	memUsed, memMax, err := getMemory()
	if err != nil {
		return nil, err
	}

	ip := GetIP()

	uptime, err := getUptime()
	if err != nil {
		return nil, err
	}

	maxCPU := runtime.NumCPU()

	health := &Health{
		CPU:     cpu,
		Disk:    diskUsed,
		ID:      hostname,
		MaxCPU:  maxCPU,
		MaxDisk: diskMax,
		MaxMem:  memMax,
		Mem:     memUsed,
		IP:      ip,
		Status:  "online",
		Uptime:  uptime,
	}

	data, err := json.MarshalIndent(health, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal data", slog.String("error", err.Error()))
		os.Exit(1)
	}

	return data, nil
}
