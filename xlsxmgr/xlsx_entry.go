package xlsxmgr

import (
	"context"

	"github.com/walleframe/walle/app"
)

// ConfigLoader 原始数据解析
type ConfigLoader interface {
	// NewContainer 新建数据指针接口
	NewContainer() interface{}
	// Check 数据检查接口.加载后校验
	Check(new interface{}) error
	// Swap 数据指针更新接口
	Swap(new interface{})
}

// ConfigAppendLoader 追加数据解析
type ConfigAppendLoader interface {
	// Parse 解析成新结构体
	Parse(basic interface{}) (new interface{}, err error)
	// Check 数据检查接口.加载后校验
	Check(new interface{}) error
	// Swap 数据指针更新接口
	Swap(new interface{})
}

// DataParser 数据解析接口
type DataParser interface {
	UnmarshalXlsxData(data []byte, v interface{}) error
}

type DataParseFunc func(data []byte, v interface{}) error

func (f DataParseFunc) UnmarshalXlsxData(data []byte, v interface{}) error {
	return f(data, v)
}

var _ DataParser = DataParseFunc(nil)

// 数据加载插件
type XlsxLoadPlugin interface {
	// Name plugin name
	Name() string
	// UnmarshalXlsxData unmarshal xlsx data to object
	UnmarshalXlsxData(data []byte, v interface{}) error
	// Start start load file
	Start(ctx context.Context, mgr *XlsxConfig, s app.Stoper) error
	// stop
	Stop(ctx context.Context)
}

var registry = pluginRegistry{
	plugins: make(map[string]XlsxLoadPlugin),
}

// RegisterXlsxPlugin register xlsx plugin to load data
var RegisterXlsxPlugin func(plugin XlsxLoadPlugin) = registry.RegisterXlsxPlugin

// GetPlugin get plugin 
var GetPlugin func(name string) (XlsxLoadPlugin, error) = registry.GetPlugin
