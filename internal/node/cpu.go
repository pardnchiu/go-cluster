package node

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

func getCpuStat() (idle, total uint64, err error) {
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

func getCpuUsage() (float64, error) {
	idle1, total1, err := getCpuStat()
	if err != nil {
		return 0, err
	}

	time.Sleep(1 * time.Second)

	idle2, total2, err := getCpuStat()
	if err != nil {
		return 0, err
	}

	idle := idle2 - idle1
	total := total2 - total1
	usage := float64(total-idle) / float64(total)

	return math.Round(usage*10000) / 10000, nil
}
