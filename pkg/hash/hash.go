// Package hash provides file hashing utilities.
package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// Algorithm represents a hash algorithm.
type Algorithm string

const (
	MD5    Algorithm = "md5"
	SHA1   Algorithm = "sha1"
	SHA256 Algorithm = "sha256"
	SHA512 Algorithm = "sha512"
)

// Result contains the hash result for a file.
type Result struct {
	File      string `json:"file"`
	Algorithm string `json:"algorithm"`
	Hash      string `json:"hash"`
	Size      int64  `json:"size_bytes"`
}

// SupportedAlgorithms returns all supported hash algorithms.
func SupportedAlgorithms() []Algorithm {
	return []Algorithm{MD5, SHA1, SHA256, SHA512}
}

// IsSupported checks if an algorithm is supported.
func IsSupported(algo string) bool {
	switch Algorithm(algo) {
	case MD5, SHA1, SHA256, SHA512:
		return true
	}
	return false
}

// newHasher returns a hash.Hash for the given algorithm.
func newHasher(algo Algorithm) (hash.Hash, error) {
	switch algo {
	case MD5:
		return md5.New(), nil
	case SHA1:
		return sha1.New(), nil
	case SHA256:
		return sha256.New(), nil
	case SHA512:
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algo)
	}
}

// File calculates the hash of a file.
func File(path string, algo Algorithm) (*Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Get file size
	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	h, err := newHasher(algo)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &Result{
		File:      path,
		Algorithm: string(algo),
		Hash:      hex.EncodeToString(h.Sum(nil)),
		Size:      stat.Size(),
	}, nil
}

// Reader calculates the hash of data from a reader.
func Reader(r io.Reader, algo Algorithm) (string, error) {
	h, err := newHasher(algo)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// String calculates the hash of a string.
func String(s string, algo Algorithm) (string, error) {
	h, err := newHasher(algo)
	if err != nil {
		return "", err
	}

	if _, err := h.Write([]byte(s)); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// Verify checks if a file matches an expected hash.
func Verify(path string, algo Algorithm, expected string) (bool, *Result, error) {
	result, err := File(path, algo)
	if err != nil {
		return false, nil, err
	}

	return result.Hash == expected, result, nil
}
