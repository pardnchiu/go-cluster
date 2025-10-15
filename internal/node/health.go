package node

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
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

func (n *Node) CheckHealth() (interface{}, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = AppName
	}

	cpu, err := cpuUsage()
	if err != nil {
		return nil, err
	}

	diskUsed, diskMax, err := disk("/")
	if err != nil {
		return nil, err
	}

	memUsed, memMax, err := memory()
	if err != nil {
		return nil, err
	}

	ip := ip()

	uptime, err := uptime()
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
		slog.Error("Failed: marshal data",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	return data, nil
}

func ip() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}

func cpuStat() (idle, total uint64, err error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)

			var user, nice, system, idleVal, iowait, irq, softirq uint64
			user, _ = strconv.ParseUint(fields[1], 10, 64)
			nice, _ = strconv.ParseUint(fields[2], 10, 64)
			system, _ = strconv.ParseUint(fields[3], 10, 64)
			idleVal, _ = strconv.ParseUint(fields[4], 10, 64)
			iowait, _ = strconv.ParseUint(fields[5], 10, 64)
			irq, _ = strconv.ParseUint(fields[6], 10, 64)
			softirq, _ = strconv.ParseUint(fields[7], 10, 64)

			idle = idleVal
			total = user + nice + system + idleVal + iowait + irq + softirq

			return idle, total, nil
		}
	}
	return 0, 0, fmt.Errorf("/proc/stat format error")
}

func cpuUsage() (float64, error) {
	idle1, total1, err := cpuStat()
	if err != nil {
		return 0, err
	}

	time.Sleep(1 * time.Second)

	idle2, total2, err := cpuStat()
	if err != nil {
		return 0, err
	}

	idle := idle2 - idle1
	total := total2 - total1
	usage := float64(total-idle) / float64(total)

	return math.Round(usage*10000) / 10000, nil
}

func disk(path string) (used, total uint64, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, err
	}

	total = stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used = total - free

	return used, total, nil
}

func memory() (used, total uint64, err error) {
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err != nil {
		return 0, 0, err
	}

	total = info.Totalram * uint64(info.Unit)
	free := info.Freeram * uint64(info.Unit)
	used = total - free
	return used, total, nil
}

func uptime() (int64, error) {
	file, err := os.Open("/proc/uptime")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var uptime float64
	fmt.Fscanf(file, "%f", &uptime)

	return int64(uptime), nil
}
