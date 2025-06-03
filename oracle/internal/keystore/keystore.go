package keystore

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

type Keystore interface {
	LoadSecret(pubkey []byte) []byte
	LoadSession(dkgUntilTimestamp int64) []byte
	LoadNonce(name string) []byte
	LoadCommitments(name string) []byte
	LoadSigningShares(name string) [][]byte
	StoreSecret(pubkey []byte, secret []byte) error
	StoreSession(dkgUntilTimestamp int64, secret []byte) error
	StoreNonce(name string, nonce []byte) error
	StoreCommitments(name string, commitments []byte) error
	StoreSigningShares(name string, pkgs [][]byte) error
	Cleanup()
}

type FileKeystore struct {
	mu       sync.Mutex
	rootPath string
}

func New(rootPath string) (Keystore, error) {
	var err error
	err = CreateDir(rootPath, "secrets", 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets dir: %w", err)
	}

	err = CreateDir(rootPath, "temp", 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	err = CreateDir(rootPath, "sessions", 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets dir: %w", err)
	}

	return &FileKeystore{rootPath: rootPath}, nil
}

func (ks *FileKeystore) load(parent string, filename string) []byte {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	filePath := filepath.Join(ks.rootPath, parent, filename)
	blob, err := os.ReadFile(filePath)
	if err != nil {
		logger.Log.Warn().Msgf("Failed to load keystore file `%s`", filePath)
		return nil
	}
	return blob
}

func (ks *FileKeystore) LoadSecret(pubkey []byte) []byte {
	fileName := hex.EncodeToString(pubkey[:32])
	return ks.load("secrets", fileName)
}

func (ks *FileKeystore) LoadSession(dkgUntilTimestamp int64) []byte {
	fileName := fmt.Sprintf("%d", dkgUntilTimestamp)
	return ks.load("sessions", fileName)
}

func (ks *FileKeystore) LoadNonce(name string) []byte {
	return ks.load("temp", "nonce_"+name)
}

func (ks *FileKeystore) LoadCommitments(name string) []byte {
	return ks.load("temp", "commitments_"+name)
}

func (ks *FileKeystore) LoadSigningShares(name string) [][]byte {
	data := ks.load("temp", "shares_"+name)
	if data == nil {
		return nil
	}
	hexString := string(data)
	lines := strings.Split(hexString, "\n")

	var packages [][]byte
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			pkg, err := hex.DecodeString(line)
			if err != nil {
				logger.Log.Error().Msgf("Failed to decode hex string `%s`", line)
				return nil
			}
			packages = append(packages, pkg)
		}
	}
	return packages
}

func (ks *FileKeystore) write(filePath string, data []byte) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	return WriteFile(filePath, data, 0o600)
}

func (ks *FileKeystore) StoreSecret(pubkey []byte, secret []byte) error {
	fileName := hex.EncodeToString(pubkey[:32])
	filePath := filepath.Join(ks.rootPath, "secrets", fileName)
	return ks.write(filePath, secret)
}

func (ks *FileKeystore) StoreSession(dkgUntilTimestamp int64, secret []byte) error {
	fileName := fmt.Sprintf("%d", dkgUntilTimestamp)
	filePath := filepath.Join(ks.rootPath, "sessions", fileName)
	return ks.write(filePath, secret)
}

func (ks *FileKeystore) StoreNonce(name string, nonce []byte) error {
	filePath := filepath.Join(ks.rootPath, "temp", "nonce_"+name)
	return ks.write(filePath, nonce)
}

func (ks *FileKeystore) StoreCommitments(name string, commitments []byte) error {
	filePath := filepath.Join(ks.rootPath, "temp", "commitments_"+name)
	return ks.write(filePath, commitments)
}

func (ks *FileKeystore) StoreSigningShares(name string, pkgs [][]byte) error {
	filePath := filepath.Join(ks.rootPath, "temp", "shares_"+name)
	// Convert each package to hex string and join with newlines
	var lines []string
	for _, pkg := range pkgs {
		lines = append(lines, hex.EncodeToString(pkg))
	}
	data := []byte(strings.Join(lines, "\n"))
	return ks.write(filePath, data)
}

func (ks *FileKeystore) Cleanup() {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	os.RemoveAll(filepath.Join(ks.rootPath, "temp"))
	CreateDir(ks.rootPath, "temp", 0o700)
}

// Creates a directory with specific access flags
// basePath: existing path (error if doesn't exist)
// subPath: path to create relative to basePath (may or may not exist)
// flags: desired access permissions
func CreateDir(
	basePath string,
	subPath string,
	flags os.FileMode,
) error {
	// Check if base path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return fmt.Errorf("base path does not exist: %s", basePath)
	}

	resultPath := filepath.Join(basePath, subPath)

	// Check if the sub path already exists
	if info, err := os.Stat(resultPath); err == nil {
		// Path exists, check if access flags match
		if info.Mode().Perm() != flags.Perm() {
			return fmt.Errorf("directory exists but permissions don't match: expected %v, got %v", flags.Perm(), info.Mode().Perm())
		}

		// Path exists and flags match, nothing to do
		return nil
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("error checking path: %v", err)
	}

	// Path doesn't exist, create it with specified flags
	if err := os.MkdirAll(resultPath, flags); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	return nil
}

// Writes data to a file with specific access flags
func WriteFile(
	filePath string,
	data []byte,
	flags os.FileMode,
) error {
	// Check if the file already exists
	if info, err := os.Stat(filePath); err == nil {
		// File exists, check if access flags match
		if info.Mode().Perm() != flags.Perm() {
			return fmt.Errorf("file exists but permissions don't match: expected %v, got %v", flags.Perm(), info.Mode().Perm())
		}

		// File exists and flags match, write data
		return os.WriteFile(filePath, data, flags)
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("error checking file: %v", err)
	}

	// Create and write file with specified flags
	if err := os.WriteFile(filePath, data, flags); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}
