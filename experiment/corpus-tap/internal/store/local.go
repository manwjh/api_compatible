package store

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LocalBlob struct {
	URI   string
	SHA256 string
	Bytes int64
}

func WriteLocalGzip(baseDir, deploymentID string, userID int, exchangeID, role string, data []byte) (LocalBlob, error) {
	if baseDir == "" {
		return LocalBlob{}, fmt.Errorf("local data dir not configured")
	}
	dt := time.Now().UTC().Format("2006-01-02")
	dir := filepath.Join(baseDir, deploymentID, fmt.Sprintf("user_id=%d", userID), "dt="+dt, exchangeID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return LocalBlob{}, err
	}
	name := role + ".json.gz"
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return LocalBlob{}, err
	}
	zw := gzip.NewWriter(f)
	n, err := zw.Write(data)
	if err != nil {
		_ = f.Close()
		return LocalBlob{}, err
	}
	if err := zw.Close(); err != nil {
		_ = f.Close()
		return LocalBlob{}, err
	}
	if err := f.Close(); err != nil {
		return LocalBlob{}, err
	}
	sum := sha256.Sum256(data)
	return LocalBlob{
		URI:    "file://" + path,
		SHA256: hex.EncodeToString(sum[:]),
		Bytes:  int64(n),
	}, nil
}
