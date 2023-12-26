package simple

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/walleframe/walle/app"
	"github.com/walleframe/walle/services/configcentra"
	"github.com/walleframe/walle/xlsxmgr"
	"go.uber.org/zap"
)

var XlsxPlugin = &LocalFileLoad{
	localPath: "./xlsxcfg/",
	fileExt:   ".json",
	Unmarshal: json.Unmarshal,
	Reader: FileReaderFunc(func(filename string) ([]byte, error) {
		// ReadFile 读取文件数据,默认直接读
		return os.ReadFile(filename)
	}),
}

func init() {
	configcentra.String(&XlsxPlugin.localPath, "xlsxcfg.simple.path", XlsxPlugin.localPath, "xlsx plugin [simple] load path")
	configcentra.String(&XlsxPlugin.fileExt, "xlsxcfg.simple.ext", XlsxPlugin.fileExt, "xlsx plugin [simple] load path")
}

type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

type FileReaderFunc func(filename string) ([]byte, error)

func (f FileReaderFunc) ReadFile(filename string) ([]byte, error) {
	if f != nil {
		return f(filename)
	}
	return nil, errors.New("no valid file reader")
}

// LocalFileLoad 本地配置文件加载插件
type LocalFileLoad struct {
	// load file path
	localPath string
	// file extern
	fileExt string
	// enable to replace by outside
	Reader FileReader
	// unmarshal func, enable to replace by outside
	Unmarshal func(data []byte, v interface{}) error
}

var _ xlsxmgr.XlsxLoadPlugin = (*LocalFileLoad)(nil)

// Name plugin name
func (l *LocalFileLoad) Name() string {
	return "simple"
}

func (l *LocalFileLoad) Start(ctx context.Context, mgr *xlsxmgr.XlsxConfig, s app.Stoper) error {
	if err := l.loadLocalConfig(mgr); err != nil {
		return err
	}
	return nil
}

func (l *LocalFileLoad) Stop(ctx context.Context) {

}

// UnmarshalXlsxData unmarshal xlsx data to object
func (l *LocalFileLoad) UnmarshalXlsxData(data []byte, v interface{}) error {
	return l.Unmarshal(data, v)
}

// loadLocalConfig 加载本地配置
func (l *LocalFileLoad) loadLocalConfig(mgr *xlsxmgr.XlsxConfig) error {
	var errCount int32

	var wg sync.WaitGroup
	mgr.Range(func(item *xlsxmgr.ConfigItem) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()

			data, err := l.loadConfigFromFile(mgr, item.GetFileName(l.fileExt))
			if err != nil {
				errCount++ // 并发.没有问题.出错了只要大于0就行
				return
			}
			if err = item.LoadConfigFromData(data); err != nil {
				errCount++
			}
		}()
		return true
	})

	wg.Wait()

	mgr.Range(func(value *xlsxmgr.ConfigItem) bool {
		value.BuildMixData()
		return true
	})

	if errCount > 0 {
		return errors.New("load local config failed")
	}
	return nil
}

// loadConfigFromFile 从本地文件中加载配置
func (l *LocalFileLoad) loadConfigFromFile(mgr *xlsxmgr.XlsxConfig, filename string) ([]byte, error) {
	localFileName := filepath.Join(l.localPath, filename)

	// 读取文件
	data, err := l.Reader.ReadFile(localFileName)
	if err != nil {
		mgr.Logger().New("configItem.loadConfigFromFile").Error("load failed, no file",
			zap.Error(err),
			zap.String("file", filename),
			zap.String("filePath", localFileName),
		)
		return nil, fmt.Errorf("file[%s] not found", filename)
	}

	return data, nil
}
