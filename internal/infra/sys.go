package infra

import (
	"syscall"
)

// DiskFree retourne pour la partition courante
func DiskFree() (float64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(".", &st); err != nil {
		return 0, err
	}
	total := float64(st.Blocks) * float64(st.Bsize)
	free := float64(st.Bavail) * float64(st.Bsize)
	return free / total * 100, nil
}

// AlertColor renvoie rouge si pourcentage < 10 %
func AlertColor(pct float64) string {
	if pct < 10 {
		return "\033[31m"
	}
	return "\033[0m"
}
