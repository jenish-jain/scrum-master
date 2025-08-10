package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SaveJSON saves data as JSON to a file
func SaveJSON(data interface{}, filepath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadJSON loads JSON data from a file
func LoadJSON(filepath string, target interface{}) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// EnsureDir ensures a directory exists
func EnsureDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// GenerateTimestamp generates a timestamp string
func GenerateTimestamp() string {
	return time.Now().Format("20060102-150405")
}

// GenerateOutputFilename generates a filename with timestamp
func GenerateOutputFilename(prefix, extension string) string {
	timestamp := GenerateTimestamp()
	return fmt.Sprintf("%s-%s.%s", prefix, timestamp, extension)
}

// GetOutputPath generates a full output path
func GetOutputPath(outputDir, filename string) string {
	return filepath.Join(outputDir, filename)
}

// FileExists checks if a file exists
func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

// ReadFile reads the entire contents of a file
func ReadFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}
