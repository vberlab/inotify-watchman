package tarutil

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/vberlabs/inotify-watchman/internal/action"
)

type ArchiveSimple struct {
	actionID         string
	ignoreArgsErrors bool
}

func NewArchiveSimple() *ArchiveSimple {
	return &ArchiveSimple{
		actionID:         "archive-simple",
		ignoreArgsErrors: false,
	}
}

func (a *ArchiveSimple) ID() string {
	return a.actionID
}

func (a *ArchiveSimple) IgnoreArgsErrors() bool {
	return a.ignoreArgsErrors
}

func (a *ArchiveSimple) ExpectedArgs() []string {
	return []string{
		"archive-destination-folder",
		"archive-name",
		"config-file-type",
		"config-path",
		"reactorName",
		"events",
	}
}

func (a *ArchiveSimple) PrepareArgs(logger *slog.Logger, args ...map[string]any) (*action.ActionArgs, error) {
	logger.Debug("ArchiveSimpleFile action PrepareArgs")
	var actionArgs = action.NewActionArgs()
	if !a.ignoreArgsErrors {
		if len(args) == 0 {
			return nil, errors.New("Arguments not passed")
		}
		if len(a.ExpectedArgs()) == 0 {
			return nil, errors.New("Expected arguments not specified")
		}
	}
	for _, item := range args {
		logger.Debug("Place ArchiveSimpleFile action args to ActionArgs struct")
		for _, argName := range a.ExpectedArgs() {
			argValue, ok := item[argName]
			if !ok {
				continue
			}
			actionArgs.Args[argName] = argValue
		}
	}
	if len(actionArgs.Args) != len(a.ExpectedArgs()) {
		return nil, errors.New("Expected arguments not fully defined")
	}
	return actionArgs, nil
}

func (a *ArchiveSimple) Execute(args action.ActionArgs, logger *slog.Logger) (string, error) {
	logger.Debug(fmt.Sprintf("Put files into tar archive %s", args.Args["archive-name"]))
	archiveFile, err := os.Create(fmt.Sprintf("%s/%s", args.Args["archive-destination-folder"], args.Args["archive-name"]))
	if err != nil {
		return "", err
	}
	defer archiveFile.Close()

	tarWriter := tar.NewWriter(archiveFile)
	events := args.Args["events"].([]fsnotify.Event)
	for _, event := range events {
		logger.Debug(fmt.Sprintf("Put %s file into archive", event.Name))
		file, err := os.Open(event.Name)
		if err != nil {
			tarWriter.Close()
			return "", err
		}

		fileInfo, err := file.Stat()
		if err != nil {
			file.Close()
			tarWriter.Close()
			return "", err
		}

		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			file.Close()
			tarWriter.Close()
			return "", err
		}

		header.Name = filepath.Base(event.Name)
		err = tarWriter.WriteHeader(header)
		if err != nil {
			file.Close()
			tarWriter.Close()
			return "", err
		}

		_, err = io.Copy(tarWriter, file)
		file.Close()

		if err != nil {
			return "", err
		}
	}
	err = tarWriter.Close()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Archive created %s", args.Args["archive-name"]), nil
}
