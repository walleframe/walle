package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/walleframe/walle/app"
	"github.com/walleframe/walle/services/configcentra"
	"github.com/walleframe/walle/util"
	"github.com/walleframe/walle/util/disk_monitor"
	"github.com/walleframe/walle/xlsxmgr"
	"github.com/walleframe/walle/zaplog"
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
	configcentra.String(&XlsxPlugin.localPath, "xlsxcfg.local.path", XlsxPlugin.localPath, "xlsx plugin [local] load path")
	configcentra.String(&XlsxPlugin.fileExt, "xlsxcfg.local.ext", XlsxPlugin.fileExt, "xlsx plugin [local] load path")
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
	//
	monitor *disk_monitor.DiskMonitor
	mgr     *xlsxmgr.XlsxConfig
	//
	cache    []string
	realPath string
}

var _ xlsxmgr.XlsxLoadPlugin = (*LocalFileLoad)(nil)

// Name plugin name
func (l *LocalFileLoad) Name() string {
	return "local"
}

func (l *LocalFileLoad) Start(ctx context.Context, mgr *xlsxmgr.XlsxConfig, s app.Stoper) (err error) {
	l.realPath, err = filepath.Abs(l.localPath)
	if err != nil {
		return fmt.Errorf("LocalFileLoad.Start convert local path to abs failed,%+v", err)
	}

	l.mgr = mgr
	l.monitor = disk_monitor.NewMonitor(
		disk_monitor.WithFileExt(l.fileExt),
		disk_monitor.WithNotifier(l),
		disk_monitor.WithLogger(mgr.Logger()),
		disk_monitor.WithPaths(l.localPath),
		disk_monitor.WithStoper(s),
	)

	return l.monitor.Start()
}

func (l *LocalFileLoad) Stop(ctx context.Context) {
	l.monitor.Stop()
}

// UnmarshalXlsxData unmarshal xlsx data to object
func (l *LocalFileLoad) UnmarshalXlsxData(data []byte, v interface{}) error {
	return l.Unmarshal(data, v)
}

func (l *LocalFileLoad) Prepare(logger *zaplog.Logger, fileName string) {
	log := l.mgr.Logger().New("LocalFileLoad.Change")
	l.cache = make([]string, 0, 64)
	// change real path
	realPath, err := filepath.Abs(l.localPath)
	if err != nil {
		log.Error("LocalFileLoad.Prepare convert local path to abs failed", zap.Error(err))
		return
	}
	l.realPath = realPath
}

func (l *LocalFileLoad) Change(logger *zaplog.Logger, fileName string, data []byte) error {
	log := l.mgr.Logger().New("LocalFileLoad.Change")
	monitorFile := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	monitorFile = strings.TrimPrefix(fileName, l.realPath)

	configItem := l.mgr.GetConfig(monitorFile)
	if configItem == nil {
		log.Debug("file update,not register, ignore", zap.String("file", fileName))
		return nil
	}
	if err := configItem.LoadConfigFromData(data); err != nil {
		log.Error("load data failed",
			zap.String("file", fileName),
			zap.String("data", util.BytesToString(data)),
			zap.Error(err),
		)
		return err
	}

	l.cache = append(l.cache, monitorFile)

	return nil
}

func (l *LocalFileLoad) Finish(logger *zaplog.Logger, fileName string) {
	for _, file := range l.cache {
		configItem := l.mgr.GetConfig(file)
		if configItem == nil {
			continue
		}
		configItem.BuildMixData()
	}
}
