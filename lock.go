package gogramps

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const lockFileName = "lock"

// ErrLocked is returned when the database is already locked by another process.
type ErrLocked struct {
	Holder string
}

func (e *ErrLocked) Error() string {
	return fmt.Sprintf("database is locked by %s", e.Holder)
}

func writeLockFile(dir string) error {
	lockPath := filepath.Join(dir, lockFileName)

	// Check if lock file already exists.
	data, err := os.ReadFile(lockPath)
	if err == nil {
		return &ErrLocked{Holder: string(data)}
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking lock file: %w", err)
	}

	text, err := lockText()
	if err != nil {
		return fmt.Errorf("generating lock text: %w", err)
	}

	if err := os.WriteFile(lockPath, []byte(text), 0o666); err != nil {
		return fmt.Errorf("writing lock file: %w", err)
	}
	return nil
}

func removeLockFile(dir string) error {
	lockPath := filepath.Join(dir, lockFileName)
	err := os.Remove(lockPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func lockText() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	u, err := user.Current()
	if err != nil {
		username := os.Getenv("USER")
		if username == "" {
			username = "unknown"
		}
		return username + "@" + hostname, nil
	}
	return u.Username + "@" + hostname, nil
}
