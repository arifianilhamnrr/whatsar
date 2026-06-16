package wa

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func acquireInstanceLock(dataDir string) (*os.File, error) {
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, err
	}

	path := filepath.Join(dataDir, ".whatsar.lock")

	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			stale, pid := isStaleLock(path)
			if stale {
				_ = os.Remove(path)
				return acquireInstanceLock(dataDir)
			}
			return nil, fmt.Errorf("whatsar sudah jalan (PID %d) — tutup proses itu dulu", pid)
		}
		return nil, err
	}

	_, _ = f.WriteString(strconv.Itoa(os.Getpid()) + "\n")
	return f, nil
}

func isStaleLock(path string) (bool, int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return true, 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return true, 0
	}
	return !processAlive(pid), pid
}

func processAlive(pid int) bool {
	if runtime.GOOS == "windows" {
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), strconv.Itoa(pid))
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func releaseInstanceLock(f *os.File, dataDir string) {
	if f == nil {
		return
	}
	path := filepath.Join(dataDir, ".whatsar.lock")
	_ = f.Close()
	_ = os.Remove(path)
}