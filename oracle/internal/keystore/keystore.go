package keystore

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

type Keystore interface {
	LoadSecret(pubkey []byte) ([]byte, error)
	StoreSecret(pubkey []byte, secret []byte) error
	LoadShare(pubkey []byte) ([]byte, error)
}

type FileKeystore struct {
	rootPath string
}

func NewKeystore(rootPath string) (Keystore, error) {
	var err error
	err = os.MkdirAll(filepath.Join(rootPath, "secrets"), 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets dir: %w", err)
	}
	err = os.MkdirAll(filepath.Join(rootPath, "temp"), 0o700)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	return &FileKeystore{rootPath: rootPath}, nil
}

func (ks *FileKeystore) LoadSecret(pubkey []byte) ([]byte, error) {
	fileName := hex.EncodeToString(pubkey[:32])
	filePath := filepath.Join(ks.rootPath, "secrets", fileName)

	secret, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load secret: %w", err)
	}
	return secret, nil
}

func (ks *FileKeystore) LoadShare(pubkey []byte) ([]byte, error) {
	fileName := hex.EncodeToString(pubkey[:32])
	filePath := filepath.Join(ks.rootPath, "temp", "share_"+fileName)

	share, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load share: %w", err)
	}
	return share, nil
}

func (ks *FileKeystore) StoreSecret(pubkey []byte, secret []byte) error {
	fileName := hex.EncodeToString(pubkey[:32])

	filePath := filepath.Join(ks.rootPath, "secrets", fileName)

	err := os.WriteFile(filePath, secret, 0o600)
	if err != nil {
		return fmt.Errorf("failed to store secret: %w", err)
	}

	return nil
}
