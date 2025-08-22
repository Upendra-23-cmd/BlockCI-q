package storage

import (
	"path/filepath"
	"fmt"
	"os"
	"time"
)

// LogStorage manages saving logs to files
type LogStorage struct {
	BaseDir string
}

// New LogStorage creates a new log storage handler
func NewLogStorage(baseDir string) *LogStorage{
	return &LogStorage{BaseDir: baseDir}
}

// SaveLog  saves logs for a given stage/step
func (ls *LogStorage) SaveLog(stage, step string , output string) (string, error){
	
	//Ensure base Directory exists
	err := os.MkdirAll(ls.BaseDir, 0775)
	if err != nil {
		return "", err
	}

	// Filename with timestamp for uniqueness
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.log", stage,step,timestamp)
	filePath := filepath.Join(ls.BaseDir, filename)

	// Write output to a file
	err = os.WriteFile(filePath, []byte(output), 0644)
	if err != nil {
		return "" , err
	}

	return filePath, nil
}

// sanitize removes special characters from step names for filenames
func sanitize(name string) string {
	clean := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' {
			clean += string(r)
		}
	}
	if clean == "" {
		return "step"
	}
	return clean
}