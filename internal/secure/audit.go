package secure

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Log(outDir, action, details string) {
	path := filepath.Join(outDir, "audit.log")
	line := fmt.Sprintf("%s | %-6s | %s\n",
		time.Now().Format("2006-01-02 15:04:05"), action, details)
	_ = os.MkdirAll(outDir, 0o755)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line)
}
