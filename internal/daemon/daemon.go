package daemon

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pardnchiu/pdcluster/internal/node"
)

type Daemon struct {
	pidPath string
}

var (
	Local = &Daemon{
		// * use local store for persistent daemon PID
		pidPath: fmt.Sprintf("/var/run/%s.pid", node.AppName),
	}
)

func (d Daemon) Init() error {
	if _, err := d.Get(); err == nil {
		return fmt.Errorf("is running")
	}
	if err := d.Add(); err != nil {
		return fmt.Errorf("failed to create")
	}
	return nil
}

func (d Daemon) Add() error {
	pid := fmt.Sprintf("%d", os.Getpid())
	return os.WriteFile(Local.pidPath, []byte(pid), 0600)
}

func (d Daemon) Remove() error {
	return os.Remove(Local.pidPath)
}

func (d Daemon) Stop() error {
	pid, err := d.Get()
	if err != nil {
		return fmt.Errorf("daemon not running")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		if err == syscall.ESRCH {
			// * process not running, cleanup file
			_ = d.Remove()
			return fmt.Errorf("daemon not found")
		}
		return err
	}

	timeout := time.After(3 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// * timeout, force kill
			if err := process.Signal(syscall.SIGKILL); err != nil {
				if err != syscall.ESRCH {
					return err
				}
			}
			_ = d.Remove()
			return nil
		case <-ticker.C:
			// * stopped, cleanup file
			if err := process.Signal(syscall.Signal(0)); err == syscall.ESRCH {
				_ = d.Remove()
				return nil
			}
		}
	}
}

func (d Daemon) Get() (int, error) {
	content, err := os.ReadFile(Local.pidPath)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, err
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// * process not running, cleanup file
		_ = d.Remove()
		return 0, err
	}

	return pid, nil
}

func (d Daemon) Status() {
	pid, err := d.Get()
	if err != nil {
		fmt.Println("Status: stopped")
		return
	}

	fmt.Printf("Status: running (PID: %d)\n", pid)
}
