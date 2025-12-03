package storage

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// Compressor handles file compression operations
type Compressor struct{}

// NewCompressor creates a new compressor
func NewCompressor() *Compressor {
	return &Compressor{}
}

// Compress compresses data using gzip
func (c *Compressor) Compress(data io.Reader) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	if _, err := io.Copy(gzWriter, data); err != nil {
		return nil, err
	}

	if err := gzWriter.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}

// Decompress decompresses gzip data
func (c *Compressor) Decompress(data io.Reader) (*bytes.Buffer, error) {
	gzReader, err := gzip.NewReader(data)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, gzReader); err != nil {
		return nil, err
	}

	return &buf, nil
}

// CalculateHash calculates SHA256 hash of data
func (c *Compressor) CalculateHash(data io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, data); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ShouldCompress determines if a file should be compressed based on type
func (c *Compressor) ShouldCompress(mimeType string) bool {
	// Don't compress already compressed formats
	compressedTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"video/",
		"audio/",
		"application/zip",
		"application/gzip",
		"application/x-rar",
	}

	for _, ct := range compressedTypes {
		if len(mimeType) >= len(ct) && mimeType[:len(ct)] == ct {
			return false
		}
	}

	return true
}
