package actions

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/lainio/err2/try"
	"github.com/vberlabs/inotify-watchman/internal/action/actions_core"
	internal_misc "github.com/vberlabs/inotify-watchman/internal/misc"
)

type WriteVaultAction struct {
	actions_core.ActionBase
}

func NewWriteVaultAction() *WriteVaultAction {
	var expectedArgs actions_core.ExpectedArgsMap = actions_core.ExpectedArgsMap{
		"files":                reflect.TypeOf([]string(nil)),
		"mount":                reflect.TypeOf(""),
		"secret-path":          reflect.TypeOf(""),
		"secret-name":          reflect.TypeOf(""),
		"with-file-extensions": reflect.TypeOf([]string(nil)),
		"token":                reflect.TypeOf(""),
		"address":              reflect.TypeOf(""),
	}
	var requiredArgs []string = []string{
		"files",
		"mount",
		"secret-path",
		"secret-name",
		"token",
		"address",
	}
	return &WriteVaultAction{
		ActionBase: actions_core.NewActionBase("write-vault", expectedArgs, requiredArgs),
	}
}

func (w *WriteVaultAction) RunAction(args actions_core.ActionArgs, logger *slog.Logger) ([]string, error) {
	var (
		//err                error
		ctx                context.Context = context.Background()
		address            string          = args.Args["address"].(string)
		token              string          = args.Args["token"].(string)
		mount              string          = args.Args["mount"].(string)
		secretPath         string          = args.Args["secret-path"].(string)
		withFileExtensions []string
		filesForWrite      []string = args.Args["files"].([]string)
		vaultClient        *api.Client
	)
	if value, found := args.Args["with-file-extensions"].([]string); found {
		withFileExtensions = value
	}
	// Find all files for write
	if len(withFileExtensions) != 0 {
		for _, file := range args.Args["files"].([]string) {
			var extraFiles []string = try.To1(w.findFilesWithExtensions(file, withFileExtensions))
			if len(extraFiles) == 0 {
				continue
			}
			filesForWrite = append(filesForWrite, extraFiles...)
			extraFiles = extraFiles[:0]
		}
	}
	// Init vault client
	vaultClient = try.To1(w.vaultClientInit(address, token))
	// Iterate over all finded files and write content to vault
	for _, file := range filesForWrite {
		var fileContent string = try.To1(internal_misc.ReadStringFromFile(file))
		var fileBaseName string = filepath.Base(file)
		w.vaultWrite(vaultClient, ctx, mount, secretPath, fileBaseName, fileContent)
	}
	return filesForWrite, nil
}

func (w *WriteVaultAction) vaultClientInit(address string, token string) (*api.Client, error) {
	var (
		cfg    *api.Config = api.DefaultConfig()
		client *api.Client
	)
	cfg.Address = address

	client = try.To1(api.NewClient(cfg))
	client.SetToken(token)
	return client, nil
}

func (w *WriteVaultAction) vaultWrite(client *api.Client, ctx context.Context, mount string, secretPath string, secretName string, value any) error {
	try.To1(
		client.KVv2(mount).Put(
			ctx,
			secretPath,
			map[string]any{
				secretName: value,
			},
		),
	)
	return nil
}

func (w *WriteVaultAction) findFilesWithExtensions(fileName string, extensions []string) ([]string, error) {
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
