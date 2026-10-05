package reactors

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
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/lainio/err2/try"
	"github.com/vberlabs/inotify-watchman/internal/reactor/reactor_core"
)

type HashChanged struct {
	reactor_core.ReactorBase
	fileHashMap map[string][]byte
}

func NewHashChanged() *HashChanged {
	var expectedArgs reactor_core.ExpectedArgsMap = reactor_core.ExpectedArgsMap{
		"events":    reflect.TypeOf([]fsnotify.Event(nil)),
		"algorithm": reflect.TypeOf(""),
	}
	return &HashChanged{
		ReactorBase: reactor_core.NewReactorBase("hash-changed", expectedArgs),
		fileHashMap: make(map[string][]byte),
	}
}

/*func (h *HashChanged) ExpectedArgs() reactor_core.ExpectedArgsMap {
	return reactor_core.ExpectedArgsMap{
		"events":    reflect.TypeOf([]fsnotify.Event(nil)),
		"algorithm": reflect.TypeOf(""), // string
	}
}*/

func (h *HashChanged) CheckCondition(logger *slog.Logger, args reactor_core.ReactorArgs) ([]string, error) {
	// Declare variables
	var (
		hashSum            []byte
		newHashSum         []byte
		hasher             hash.Hash
		events             []fsnotify.Event
		fileName           string
		algorithm          string
		filesConditionTrue []string
	)
	// Get hash algorithm from config
	algorithm = args.Args["algorithm"].(string)
	switch algorithm {
	case "sha256":
		logger.Debug("Using sha256 algorithm")
		hasher = sha256.New()
	case "md5":
		logger.Debug("Using md5 algorithm")
		hasher = md5.New()
	default:
		if h.IgnoreErrors() {
			logger.Warn(fmt.Sprintf("Get unsupported algorithm name %s. Using sha256.", algorithm))
			hasher = sha256.New()
		} else {
			return nil, errors.New(fmt.Sprintf("Unsupported algorithm %s", algorithm))
		}
	}
	// Check conditions for passed events
	events = args.Args["events"].([]fsnotify.Event)
	for _, event := range events {
		fileName = event.Name
		hashSum = h.getHashFromMap(fileName)
		if hashSum == nil {
			hashSum = h.computeFileHash(hasher, fileName)
			goto writeHash
		}
		newHashSum = h.computeFileHash(hasher, fileName)
		if !h.compareHash(hashSum, newHashSum) {
			hashSum = newHashSum
			goto writeHash
		}
		continue
	writeHash:
		if h.writeHashMap(fileName, hashSum) {
			logger.Debug(fmt.Sprintf("Hash for file %s replaced", fileName))
		} else {
			logger.Debug(fmt.Sprintf("File %s hash does not exists, write new", fileName))
		}
		filesConditionTrue = append(filesConditionTrue, fileName)
	}
	return filesConditionTrue, nil
}

func (h *HashChanged) compareHash(oldHash []byte, newHash []byte) bool {
	if !bytes.Equal(oldHash, newHash) {
		return false
	}
	return true
}

func (h *HashChanged) getHashFromMap(fileName string) []byte {
	var (
		hashFromMap []byte
		exist       bool
	)
	hashFromMap, exist = h.fileHashMap[fileName]
	if !exist {
		return nil
	}
	return hashFromMap
}

func (h *HashChanged) writeHashMap(fileName string, newHash []byte) (replaced bool) {
	_, replaced = h.fileHashMap[fileName]

	h.fileHashMap[fileName] = newHash

	return replaced
}

func (h *HashChanged) computeFileHash(hasher hash.Hash, fileName string) []byte {
	var fD *os.File = try.To1(os.Open(fileName))
	defer fD.Close()
	try.To1(io.Copy(hasher, fD))
	return hasher.Sum(nil)
}
