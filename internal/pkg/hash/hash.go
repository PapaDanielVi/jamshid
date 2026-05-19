package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
)

// DirHash returns a short hex hash (first 8 chars of SHA-256) of the given directory path.
func DirHash(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	h := sha256.Sum256([]byte(abs))
	return hex.EncodeToString(h[:])[:8]
}
