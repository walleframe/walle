package xlsxmgr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/walleframe/walle/app"
	"github.com/walleframe/walle/zaplog"
	"go.uber.org/zap"
)

// XlsxConfig excel 配置文件管理
type XlsxConfig struct {
	mutex  sync.Mutex     // 配置互斥锁
	plugin XlsxLoadPlugin // 加载插件

	logger *zaplog.Logger
	ctx    context.Context
	cancel func()

	data sync.Map // 配置数据
}

func NewXlsxConfigManager() *XlsxConfig {
	return &XlsxConfig{}
}

////////////////////////////////////////////////////////////////////////////////
// 外部管理函数

// Init 初始化xlsx管理器
func (mgr *XlsxConfig) Init(plugin XlsxLoadPlugin, logger *zaplog.Logger) (err error) {
	mgr.plugin = plugin
	mgr.logger = logger.Named("xlsxmgr")
	mgr.ctx, mgr.cancel = context.WithCancel(context.Background())
	return nil
}

// UpdateConfig 加载配置
func (mgr *XlsxConfig) UpdateConfig(s app.Stoper) error {
	mgr.mutex.Lock() // 保证同一时间. 只运行一次
	defer mgr.mutex.Unlock()

	return mgr.plugin.Start(mgr.ctx, mgr, s)
}

// Stop 停止监听变动
func (mgr *XlsxConfig) Stop() {
	mgr.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	mgr.plugin.Stop(ctx)
}

////////////////////////////////////////////////////////////////////////////////
// 内部插件使用函数

// GetConfig 获取配置管理项
func (mgr *XlsxConfig) GetConfig(filename string) *ConfigItem {
	v, ok := mgr.data.Load(filename)
	if !ok {
		return nil
	}
	item, ok := v.(*ConfigItem)
	if !ok {
		return nil
	}
	return item
}

// Range 遍历已注册配置
func (mgr *XlsxConfig) Range(cb func(value *ConfigItem) bool) {
	mgr.data.Range(func(key, value interface{}) bool {
		item := value.(*ConfigItem)
		return cb(item)
	})
}

// Logger 获取日志接口
func (mgr *XlsxConfig) Logger() *zaplog.Logger {
	return mgr.logger
}

////////////////////////////////////////////////////////////////////////////////
// 导出外部接口

// RegAutoConfig 注册自动加载配置
func (mgr *XlsxConfig) RegAutoConfig(basename, fromXlsx, fromSheet string, parser DataParser, loader ConfigLoader) {
	item := mgr.GetConfig(basename)
	if item != nil {
		panic(fmt.Sprintf("%s(%s %s) had registered already, last %s %s:\n%s", basename,
			fromXlsx, fromSheet, fromXlsx, fromSheet,
			zap.StackSkip("", 1).String,
		))
	}
	item = &ConfigItem{
		basename:  basename,
		fromXlsx:  fromXlsx,
		fromSheet: fromSheet,
		parser:    parser,
		loader:    loader,
		mgr:       mgr,
	}
	mgr.data.Store(basename, item)
	return
}

// RegAppendConfig 追加数据解析
func (mgr *XlsxConfig) RegAppendConfig(basename, tag string, loader ConfigAppendLoader) {
	item := mgr.GetConfig(basename)
	if item == nil {
		panic(fmt.Sprintf("AppendConfig %s depend %s must register first", tag, basename))
	}
	item.externs = append(item.externs, externLoaderCache{
		tag:    tag,
		loader: loader,
	})
}

// RegMixConfig 混合数据接口
func (mgr *XlsxConfig) RegMixConfig(tag string, mix func() error, basename ...string) {
	for _, basename := range basename {
		item := mgr.GetConfig(basename)
		if item == nil {
			panic(fmt.Sprintf("MixConfig %s depend %s must register first", tag, basename))
		}
		item.mixConfig = append(item.mixConfig, mixLoaderCache{
			tag:          tag,
			buildMixData: mix,
		})
	}
}
