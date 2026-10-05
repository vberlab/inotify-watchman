package actions

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/lainio/err2/try"
	"github.com/vberlabs/inotify-watchman/internal/action/actions_core"
)

type ArchiveFileAction struct {
	actions_core.ActionBase
}

func NewArchiveFileAction() *ArchiveFileAction {
	var expectedArgs actions_core.ExpectedArgsMap = actions_core.ExpectedArgsMap{
		"files":                reflect.TypeOf([]string(nil)),
		"with-file-extensions": reflect.TypeOf([]string(nil)),
		"archive-destination":  reflect.TypeOf(""),
	}
	var requiredArgs []string = []string{
		"files",
		"archive-destination",
	}
	return &ArchiveFileAction{
		ActionBase: actions_core.NewActionBase("archive-file", expectedArgs, requiredArgs),
	}
}

func (a *ArchiveFileAction) RunAction(args actions_core.ActionArgs, logger *slog.Logger) ([]string, error) {
	var (
		archiveDestination  string = args.Args["archive-destination"].(string)
		archiveName         string
		archiveCreatedNames []string
		filesWantToArchive  []string = args.Args["files"].([]string)
		withFileExtensions  []string
		filesWillBeArchive  []string
		filesWithExtensions []string
	)
	if value, found := args.Args["with-file-extensions"].([]string); found {
		logger.Debug(fmt.Sprintf("with-file-extensions argument set %v", value))
		withFileExtensions = value
	}

	for _, file := range filesWantToArchive {
		logger.Debug("Archive file %s", file)
		archiveName = makeArchiveName(file, archiveDestination)
		filesWillBeArchive = append(filesWillBeArchive, file)
		if withFileExtensions != nil {
			filesWithExtensions = try.To1(findWithFiles(file, withFileExtensions))
		}
		filesWillBeArchive = append(filesWillBeArchive, filesWithExtensions...)
		try.To(createArchive(filesWillBeArchive, archiveName))
		filesWillBeArchive = filesWillBeArchive[:0]
		archiveCreatedNames = append(archiveCreatedNames, archiveName)
	}
	return archiveCreatedNames, nil
}

func makeArchiveName(file string, destination string) string {
	var (
		fileBaseName             = filepath.Base(file)
		currentExtension         = filepath.Ext(file)
		fileNameWithoutExtension = strings.TrimSuffix(fileBaseName, currentExtension)
	)
	return filepath.Join(destination, fileNameWithoutExtension+".tar")
}

func createArchive(files []string, destination string) error {
	var (
		archiveFile    *os.File
		err            error
		tarWriter      *tar.Writer
		tmpDestination string = "/tmp/inotify-watchman/archive-file.tar.tmp"
	)

	archiveFile, err = os.Create(tmpDestination)
	if err != nil {
		return fmt.Errorf(
			"create archive %q: %w",
			destination,
			err,
		)
	}
	// Close file if error accured
	defer func() {
		var closeErr error
		closeErr = archiveFile.Close()
		if err == nil {
			err = closeErr
		}

		if err != nil {
			_ = os.Remove(tmpDestination)
		}
	}()

	tarWriter = tar.NewWriter(archiveFile)
	// Write tar close blocks
	defer func() {
		var closeErr error
		closeErr = tarWriter.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for _, fileName := range files {
		err = addFileToArchive(tarWriter, fileName)
		if err != nil {
			return fmt.Errorf(
				"add file %q to archive %w",
				fileName,
				err,
			)
		}
	}
	err = os.Rename(tmpDestination, destination)
	if err != nil {
		os.Remove(tmpDestination)
		return err
	}
	return nil
}

func addFileToArchive(tarWriter *tar.Writer, fileName string) error {
	var (
		fileInfo  os.FileInfo
		file      *os.File
		tarHeader *tar.Header
	)

	fileInfo = try.To1(os.Stat(fileName))
	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("not a regular file: %q", fileName)
	}

	file = try.To1(os.Open(fileName))
	defer file.Close()

	tarHeader = try.To1(tar.FileInfoHeader(fileInfo, ""))
	tarHeader.Name = filepath.ToSlash(filepath.Base(fileName))
	try.To(tarWriter.WriteHeader(tarHeader))
	_ = try.To1(io.Copy(tarWriter, file))
	return nil
}

func findWithFiles(fileName string, extensions []string) ([]string, error) {
	var (
		directory            string = filepath.Dir(fileName)
		baseName             string = filepath.Base(fileName)
		currentExtension     string = filepath.Ext(baseName)
		nameWithoutExtension string = strings.TrimSuffix(baseName, currentExtension)
		relatedFile          string
		info                 os.FileInfo
		err                  error
		files                []string
	)
	for _, extension := range extensions {
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
		relatedFile = filepath.Join(directory, nameWithoutExtension+extension)
		info, err = os.Stat(relatedFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		if info.IsDir() {
			continue
		}
		files = append(files, relatedFile)
	}
	return files, nil
}
