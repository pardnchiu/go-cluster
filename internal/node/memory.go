package node

import (
	"syscall"
)

func getMemory() (used, total uint64, err error) {
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err != nil {
		return 0, 0, err
	}

	total = info.Totalram * uint64(info.Unit)
	free := info.Freeram * uint64(info.Unit)
	used = total - free

	return used, total, nil
}
