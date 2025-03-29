package keystore

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Keystore interface {
	LoadSecret(pubkey []byte) []byte
	LoadSession(dkgUntilTimestamp uint64) []byte
	LoadNonce(name string) []byte
	LoadCommitments(name string) []byte
	LoadSigningShares(name string) [][]byte
	StoreSecret(pubkey []byte, secret []byte) error
	StoreSession(dkgUntilTimestamp uint64, secret []byte) error
	StoreNonce(name string, nonce []byte) error
	StoreCommitments(name string, commitments []byte) error
	StoreSigningShares(name string, pkgs [][]byte) error
	Cleanup()
}

type FileKeystore struct {
	rootPath string
}

func New(rootPath string) (Keystore, error) {
	var err error
	err = os.MkdirAll(filepath.Join(rootPath, "secrets"), 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets dir: %w", err)
	}

	err = os.MkdirAll(filepath.Join(rootPath, "temp"), 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	err = os.MkdirAll(filepath.Join(rootPath, "sessions"), 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets dir: %w", err)
	}

	return &FileKeystore{rootPath: rootPath}, nil
}

func (ks *FileKeystore) load(parent string, filename string) []byte {
	filePath := filepath.Join(ks.rootPath, parent, filename)
	blob, err := os.ReadFile(filePath)
	if err != nil {
		// TODO add error log
		return nil
	}
	return blob
}

func (ks *FileKeystore) LoadSecret(pubkey []byte) []byte {
	fileName := hex.EncodeToString(pubkey[:32])
	return ks.load("secrets", fileName)
}

func (ks *FileKeystore) LoadSession(dkgUntilTimestamp uint64) []byte {
	fileName := strconv.FormatUint(dkgUntilTimestamp, 10)
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
				// TODO log error
				return nil
			}
			packages = append(packages, pkg)
		}
	}
	return packages
}

func (ks *FileKeystore) StoreSecret(pubkey []byte, secret []byte) error {
	fileName := hex.EncodeToString(pubkey[:32])
	filePath := filepath.Join(ks.rootPath, "secrets", fileName)
	return os.WriteFile(filePath, secret, 0o600)
}

func (ks *FileKeystore) StoreSession(dkgUntilTimestamp uint64, secret []byte) error {
	fileName := strconv.FormatUint(dkgUntilTimestamp, 10)
	filePath := filepath.Join(ks.rootPath, "sessions", fileName)
	return os.WriteFile(filePath, secret, 0o600)
}

func (ks *FileKeystore) StoreNonce(name string, nonce []byte) error {
	filePath := filepath.Join(ks.rootPath, "temp", "nonce_"+name)
	return os.WriteFile(filePath, nonce, 0o600)
}

func (ks *FileKeystore) StoreCommitments(name string, commitments []byte) error {
	filePath := filepath.Join(ks.rootPath, "temp", "commitments_"+name)
	return os.WriteFile(filePath, commitments, 0o600)
}

func (ks *FileKeystore) StoreSigningShares(name string, pkgs [][]byte) error {
	filePath := filepath.Join(ks.rootPath, "temp", "shares_"+name)
	// Convert each package to hex string and join with newlines
	var lines []string
	for _, pkg := range pkgs {
		lines = append(lines, hex.EncodeToString(pkg))
	}
	data := []byte(strings.Join(lines, "\n"))
	return os.WriteFile(filePath, data, 0o600)
}

func (ks *FileKeystore) Cleanup() {
	os.RemoveAll(filepath.Join(ks.rootPath, "temp"))
	os.MkdirAll(filepath.Join(ks.rootPath, "temp"), 0o700)
}
