package node

import (
	"fmt"
	"os"
)

func getUptime() (int64, error) {
	file, err := os.Open("/proc/uptime")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var uptime float64
	fmt.Fscanf(file, "%f", &uptime)

	return int64(uptime), nil
}
