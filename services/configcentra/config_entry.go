package configcentra

import (
	"time"

	"log"

	"github.com/walleframe/walle/app"
	"github.com/walleframe/walle/app/bootstrap"
)

type ConfigCentra interface {
	// backend service interface
	Init(s app.Stoper) (err error)
	Start(s app.Stoper) error
	Stop()
	Finish()
	// register custom config value
	RegisterConfig(v ConfigValue, ntf []ConfigUpdateNotify)
	// watch config update
	WatchConfigUpdate(ntf []ConfigUpdateNotify)

	// object support
	UseObject() bool
	SetObject(key string, doc string, obj interface{})
	GetObject(key string, obj interface{}) (err error)

	// static value interface
	SetDefault(key string, doc string, value interface{})
	GetString(key string) (string, error)
	GetBool(key string) (bool, error)
	GetInt(key string) (int, error)
	GetInt32(key string) (int32, error)
	GetInt64(key string) (int64, error)
	GetUint(key string) (uint, error)
	GetUint16(key string) (uint16, error)
	GetUint32(key string) (uint32, error)
	GetUint64(key string) (uint64, error)
	GetFloat64(key string) (float64, error)
	GetTime(key string) (time.Time, error)
	GetDuration(key string) (time.Duration, error)
	GetIntSlice(key string) ([]int, error)
	GetStringSlice(key string) ([]string, error)
}

// 配置中心后端实现接口
var ConfigCentraBackend ConfigCentra

// ConfigValue 配置项 配置值
type ConfigValue interface {
	SetDefaultValue(vcfg ConfigCentra)
	RefreshValue(vcfg ConfigCentra) error
}

type ConfigUpdateNotify func(ConfigCentra)
type FlagNotify func() error

// 配置中心
var gConfigManager = NewConfigService()

func init() {
	// config centra must start first of all.
	bootstrap.RegisterServiceByPriority(-1, gConfigManager) // config manager (load config from file)
}

// RegisterFlagNotify 注册flag处理
func RegisterFlagNotify(ntf FlagNotify) {
	if gConfigManager.start.Load() {
		log.Panic("service already start, CAN NOT register flag notify any more.")
	}
	gConfigManager.flags = append(gConfigManager.flags, ntf)
}

// RegisterConfig 注册配置
func RegisterConfig(cfg ConfigValue, ntf ...ConfigUpdateNotify) {
	if gConfigManager.start.Load() {
		log.Panic("service already start, CAN NOT register config any more.")
	}
	gConfigManager.values = append(gConfigManager.values, cacheValue{
		cv:  cfg,
		ntf: ntf,
	})
}

// WatchConfigUpdate 监控配置更新
func WatchConfigUpdate(ntf ConfigUpdateNotify) {
	if gConfigManager.start.Load() {
		log.Panic("service already start, CAN NOT watch config any more.")
	}
	gConfigManager.updates = append(gConfigManager.updates, ntf)
}
