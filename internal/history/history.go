package history

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const timestampFormat = "20060102T150405Z"

// SaveSnapshot copies the current normalized snapshot into historyDir using a UTC timestamp filename.
func SaveSnapshot(currentPath, historyDir string) (string, error) {
	if historyDir == "" {
		historyDir = filepath.Join(".cvx", "history")
	}
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		return "", fmt.Errorf("create history directory %s: %w", historyDir, err)
	}

	data, err := os.ReadFile(currentPath)
	if err != nil {
		return "", fmt.Errorf("read current snapshot %s: %w", currentPath, err)
	}

	savedPath := filepath.Join(historyDir, time.Now().UTC().Format(timestampFormat)+".json")
	if err := os.WriteFile(savedPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write history snapshot %s: %w", savedPath, err)
	}
	return savedPath, nil
}
