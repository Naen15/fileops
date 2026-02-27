package secure

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Lock(target, outDir string) (string, error) {
	lock := filepath.Join(outDir, filepath.Base(target)+".lock")
	if _, err := os.Stat(lock); err == nil {
		return lock, fmt.Errorf("déjà verrouillé (%s)", lock)
	}
	if err := os.WriteFile(lock, []byte(time.Now().Format(time.RFC3339)), 0o644); err != nil {
		return lock, err
	}
	return lock, nil
}
func Unlock(target, outDir string) (string, error) {
	lock := filepath.Join(outDir, filepath.Base(target)+".lock")
	if err := os.Remove(lock); err != nil {
		return lock, err
	}
	return lock, nil
}

func MakeReadOnly(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := info.Mode().Perm() & 0o555
	return os.Chmod(path, mode)
}

func IsWritable(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err == nil {
		f.Close()
		return true
	}
	return false
}
