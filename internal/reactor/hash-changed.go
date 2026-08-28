package reactor

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"log/slog"
	"os"
)

type HashChanged struct {
	reactorID   string
	fileHashMap map[string][]bytes
}

func (h *HashChanged) ID() string {
	return h.reactorID
}

func (h *HashChanged) CheckCondition(args map[string]any) (bool, error) {
	var hashSum []byte
	var newHashSum []byte

	// Check function args, need a file name
	fileName, ok := args["file"].(string)
	if !ok {
		var errorKeyNotFound error = errors.New("Missing file name in HashChanged reactor check condition function")
		slog.Error(errorKeyNotFound.Error())
		return false, errorKeyNotFound
	}
	// Open file and compute hash
	fileD, err := os.Open(fileName)
	if err != nil {
		return false, err
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, fileD); err != nil {
		slog.Error("Hash computing of file %s fail", fileName)
		return false, err
	}
	if hashSum, ok = h.fileHashMap[fileName]; !ok {
		h.fileHashMap[fileName] = hasher.Sum(nil)
		return true, nil
	}
	if !bytes.Equal(hashSum, newHashSum) {
		h.fileHashMap[fileName] = newHashSum
		return true, nil
	}
	return false, nil
}
