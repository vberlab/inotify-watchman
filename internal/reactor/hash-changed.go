package reactor

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"log/slog"
	"os"
)

func NewHashChanged() *HashChanged {
	return &HashChanged{
		reactorID:   "hash-changed",
		fileHashMap: make(map[string][]byte),
	}
}

type HashChanged struct {
	reactorID   string
	fileHashMap map[string][]byte
}

func (h *HashChanged) ID() string {
	return h.reactorID
}

func (h *HashChanged) CheckCondition(cfgArgs map[string]any, fileName string, logger *slog.Logger) (bool, error) {
	var hashSum []byte
	var newHashSum []byte
	var hasher hash.Hash

	algorithm, ok := cfgArgs["algorithm"].(string)
	if !ok {
		return false, errors.New("Algorithm argument not defined")
	}
	switch algorithm {
	case "sha256":
		logger.Debug(fmt.Sprintf("%s-%s: Using sha256 algorithm", h.ID(), fileName))
		hasher = sha256.New()
	case "md5":
		logger.Debug(fmt.Sprintf("%s-%s:Using md5 algoritm", h.ID(), fileName))
		hasher = md5.New()
	default:
		logger.Error(fmt.Sprintf("%s: Unsupported algorithm %s", fileName, algorithm))
		return false, errors.New("Unsupported algorithm")
	}

	// Open file and compute hash
	fileD, err := os.Open(fileName)
	if err != nil {
		return false, err
	}

	if _, err := io.Copy(hasher, fileD); err != nil {
		logger.Error(fmt.Sprintf("%s-%s:Hash computing of file %s fail", h.ID(), fileName, fileName))
		return false, err
	}
	hashSum, ok = h.fileHashMap[fileName]
	if !ok {
		h.fileHashMap[fileName] = hasher.Sum(nil)
		logger.Debug(fmt.Sprintf("%s-%s: File %s hash missing. New hash %s was added.", h.ID(), fileName, fileName, h.fileHashMap[fileName]))
		return true, nil
	} else {
		newHashSum = hasher.Sum(nil)
	}
	if !bytes.Equal(hashSum, newHashSum) {
		h.fileHashMap[fileName] = newHashSum
		logger.Debug(fmt.Sprintf("%s-%s:Hash of file %s was changed. New hash %s was added.", h.ID(), fileName, fileName, h.fileHashMap[fileName]))
		return true, nil
	}
	return false, nil
}
