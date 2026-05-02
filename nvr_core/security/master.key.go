package security

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
)

// LoadOrCreateMasterKey reads the 32-byte key from the specified path.
// If the file is missing, it automatically generates a secure key and locks it down.
func LoadOrCreateMasterKey(path string) ([]byte, error) {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		fmt.Printf("Master key not found at %s. Generating a new secure key...\n", path)
		return generateAndSaveKey(path)
	} else if err != nil {
		return nil, fmt.Errorf("failed to access key file: %w", err)
	}

	// Cross-platform UNIX permission check (works on macOS and Linux)
	// We expect 0400 (read-only by owner) or 0600 (read-write by owner)
	perm := info.Mode().Perm()
	if perm != 0400 && perm != 0600 {
		return nil, fmt.Errorf("FATAL: key file permissions (%o) are too open. Please run: chmod 0400 %s", perm, path)
	}

	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read key file: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("FATAL: master key is corrupted. Expected 32 bytes, found %d", len(key))
	}

	return key, nil
}

// generateAndSaveKey creates a cryptographically secure 32-byte key and writes it to disk.
func generateAndSaveKey(path string) ([]byte, error) {
	// Ensure the parent directory exists with safe permissions (0700)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("could not create directory for key: %w", err)
	}

	// Generate 32 bytes of cryptographic randomness
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("could not generate random key data: %w", err)
	}

	// Write to disk with strictly locked down permissions (0400)
	if err := os.WriteFile(path, key, 0400); err != nil {
		return nil, fmt.Errorf("could not save key file: %w", err)
	}

	return key, nil
}