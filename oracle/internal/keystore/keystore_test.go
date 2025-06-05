package keystore

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TmpRootPath() string {
	return filepath.Join(os.TempDir(), "oracle_tmp_test")
}

func ClearTmp(t *testing.T) {
	tmpDir := TmpRootPath()

	_, err := os.Stat(tmpDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(tmpDir, FileMaskOwnerRWX); err != nil {
				t.Errorf("failed to create directory: %v", err)
				return
			}
			return
		}

		t.Errorf("%v", err)
		return
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	for _, entry := range entries {
		path := filepath.Join(tmpDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			t.Errorf("%v", fmt.Errorf("removing %s: %w", path, err))
			return
		}
	}
}

func VerifyExists(filePath string, flags os.FileMode, t *testing.T) {
	tmpDir := TmpRootPath()
	path := filepath.Join(tmpDir, filePath)

	if info, err := os.Stat(path); err == nil {
		// File exists, check if access flags match
		if info.Mode().Perm() != flags.Perm() {
			t.Errorf("file '%s' exists but permissions don't match: expected %v, got %v", path, flags.Perm(), info.Mode().Perm())
			return
		}
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		t.Errorf("error checking file: %v", err)
		return
	} else {
		t.Errorf("file '%s' does not exist", path)
		return
	}
}

func ClearAndCreate(t *testing.T) Keystore {
	ClearTmp(t)
	ks, err := New(TmpRootPath())

	if err != nil {
		t.Errorf("%v", err)
		return nil
	}

	// Verify that the folders exist (created)
	VerifyExists("secrets", FileMaskOwnerRWX, t)
	VerifyExists("temp", FileMaskOwnerRWX, t)
	VerifyExists("sessions", FileMaskOwnerRWX, t)

	return ks
}

func ChangePermissions(filePath string, flags os.FileMode, t *testing.T) {
	tmpDir := TmpRootPath()
	path := filepath.Join(tmpDir, filePath)

	err := os.Chmod(path, flags)
	if err != nil {
		t.Errorf("Error changing permissions: %v", err)
		return
	}
}

func TestKeystoreFolderCreation_Success(t *testing.T) {
	ClearAndCreate(t)
}

func TestKeystoreFolderCreationWithExistedFolders_Success(t *testing.T) {
	// Clear and create
	ClearAndCreate(t)

	// Create again
	_, err := New(TmpRootPath())

	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// Verify that the folders exist (created)
	VerifyExists("secrets", FileMaskOwnerRWX, t)
	VerifyExists("temp", FileMaskOwnerRWX, t)
	VerifyExists("sessions", FileMaskOwnerRWX, t)
}

func TestKeystoreFolderCreationWithExistedFoldersWrongPermissions_Success(t *testing.T) {
	// Clear and create
	ClearAndCreate(t)

	// Change permissions
	ChangePermissions("secrets", FileMaskOwnerGRoupOtherRWX, t)
	ChangePermissions("temp", FileMaskOwnerGRoupOtherRWX, t)
	ChangePermissions("sessions", FileMaskOwnerGRoupOtherRWX, t)

	// Create again
	_, err := New(TmpRootPath())

	if err == nil {
		t.Errorf("Expected error 'directory exists but permissions don't match'")
		return
	}
}

func TestKeystoreWriteFile(t *testing.T) {
	// Clear and create
	ks := ClearAndCreate(t)

	// Write file
	pubkey := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	pubkeyData, err := hex.DecodeString(pubkey)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = ks.StoreSecret(pubkeyData, []byte("test payload"))
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// Change permissions
	ChangePermissions(filepath.Join("secrets", pubkey), FileMaskOwnerGRoupOtherRWX, t)
	VerifyExists(filepath.Join("secrets", pubkey), FileMaskOwnerGRoupOtherRWX, t)

	// Write file
	err = ks.StoreSecret(pubkeyData, []byte("test payload"))
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	VerifyExists(filepath.Join("secrets", pubkey), FileMaskOwnerRW, t)
}
